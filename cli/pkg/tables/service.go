package tables

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/rio/cli/cmd/util"

	webhookv1 "github.com/rancher/gitwatcher/pkg/apis/gitwatcher.cattle.io/v1"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/modules/service/controllers/service/populate/serviceports"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/controllers/pkg"
	"github.com/rancher/rio/pkg/riofile/stringers"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func NewService(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{.Obj | id}}"},
		{"IMAGE", "{{.Obj | image}}"},
		{"ENDPOINT", "{{.Obj | formatEndpoint }}"},
		{"PORTS", "{{.Obj | formatPorts}}"},
		{"SCALE", "{{.Obj | scale}}"},
		{"WEIGHT", "{{.Obj | formatWeight}}"},
		{"CREATED", "{{.Obj.CreationTimestamp | ago}}"},
		{"DETAIL", "{{serviceDetail .Data.Service .Data.Pods .Data.GitWatcher}}"},
	}, cfg)

	writer.AddFormatFunc("image", FormatImage)
	writer.AddFormatFunc("scale", formatRevisionScale)
	writer.AddFormatFunc("formatPorts", formatPorts)
	writer.AddFormatFunc("formatEndpoint", formatEndpoint)
	writer.AddFormatFunc("formatWeight", formatWeight)
	writer.AddFormatFunc("serviceDetail", serviceDetail)

	return &tableWriter{
		writer: writer,
	}
}

func formatPorts(data interface{}) string {
	s, ok := data.(*v1.Service)
	if !ok {
		return ""
	}

	buf := strings.Builder{}
	for _, port := range serviceports.ContainerPorts(s) {
		cp := stringers.ContainerPortStringer{
			ContainerPort: port,
		}
		if buf.Len() > 0 {
			buf.WriteString(",")
		}
		s := cp.MaybeString()
		if str, ok := s.(string); ok {
			buf.WriteString(str)
		}
	}

	return buf.String()
}

func formatEndpoint(data interface{}) string {
	s, ok := data.(*v1.Service)
	if !ok {
		return ""
	} else if len(s.Status.Endpoints) > 0 {
		endpoints := util.NormalizingEndpoints(s.Status.Endpoints, "")
		return strings.Join(endpoints, ",")
	}
	return ""
}

func formatWeight(data interface{}) string {
	s, ok := data.(*v1.Service)
	if !ok {
		return ""
	}
	if len(s.Status.Endpoints) == 0 {
		return ""
	}
	if s.Status.ComputedWeight != nil {
		return fmt.Sprintf("%s%%", strconv.Itoa(*s.Status.ComputedWeight))
	}

	return "0%"
}

func serviceDetail(data interface{}, pods []*corev1.Pod, gitwatcher *webhookv1.GitWatcher) string {
	s, ok := data.(*v1.Service)
	if !ok || s == nil {
		return ""
	}

	if s.Spec.Template {
		return "Build Template"
	}

	buffer := strings.Builder{}

	pd := ""
	for _, pod := range pods {
		if pod == nil {
			continue
		}
		pd = pkg.PodDetail(pod)
		if pd != "" {
			break
		}
	}
	if pd == "" && !s.Status.DeploymentReady {
		buffer.WriteString(s.Name + ": ")
		pd = "not ready"
	}
	buffer.WriteString(pd)

	if waitingOnBuild(s) {
		if gitwatcher != nil && webhookv1.GitWebHookReceiverConditionRegistered.IsFalse(gitwatcher) {
			message := webhookv1.GitWebHookReceiverConditionRegistered.GetMessage(gitwatcher)
			reason := webhookv1.GitWebHookReceiverConditionRegistered.GetReason(gitwatcher)
			return fmt.Sprintf("Failed to watch git repo: %s(%s)", message, reason)
		}
		if riov1.ServiceConditionImageReady.IsFalse(s) {
			if buffer.Len() > 0 {
				buffer.WriteString("; ")
			}
			buffer.WriteString(s.Name)
			buffer.WriteString(" build failed: ")
			buffer.WriteString(riov1.ServiceConditionImageReady.GetMessage(s))
		} else if !riov1.ServiceConditionImageReady.IsTrue(s) {
			if buffer.Len() > 0 {
				buffer.WriteString("; ")
			}
			buffer.WriteString(s.Name)
			buffer.WriteString(" waiting on build")
		}
	}
	if buffer.Len() > 0 {
		return buffer.String()
	}

	for _, con := range s.Status.Conditions {
		if con.Status != corev1.ConditionTrue {
			return fmt.Sprintf("%s: %s(%s)", con.Type, con.Message, con.Reason)
		}
	}

	return ""
}

func formatRevisionScale(data interface{}) (string, error) {
	switch v := data.(type) {
	case *v1.Service:
		avail, unavail := 0, 0
		if v.Status.ScaleStatus != nil {
			avail = v.Status.ScaleStatus.Available
			unavail = v.Status.ScaleStatus.Unavailable
		}
		return FormatScale(v.Status.ComputedReplicas, avail, unavail)
	case *appsv1.DaemonSet:
		scale := int(v.Status.DesiredNumberScheduled)
		return FormatScale(&scale,
			int(v.Status.NumberAvailable),
			int(v.Status.NumberUnavailable))
	case *appsv1.Deployment:
		var scale *int
		if v.Spec.Replicas != nil {
			iScale := int(*v.Spec.Replicas)
			scale = &iScale
		}
		return FormatScale(scale,
			int(v.Status.AvailableReplicas),
			int(v.Status.UnavailableReplicas))
	}
	return "", nil
}

func FormatScale(scale *int, available, unavailable int) (string, error) {
	scaleNum := 1
	if scale != nil {
		scaleNum = *scale
	}

	scaleStr := strconv.Itoa(scaleNum)

	if scaleNum == -1 {
		return strconv.Itoa(available), nil
	}

	if unavailable == 0 {
		return scaleStr, nil
	}

	var prefix string
	percentage := ""
	ready := available
	if scaleNum > 0 {
		percentage = fmt.Sprintf(" %d%%", (ready*100)/scaleNum)
	}

	prefix = fmt.Sprintf("%d/", ready)

	return fmt.Sprintf("%s%d%s", prefix, scaleNum, percentage), nil
}

func getServiceImage(s *v1.Service) string {
	if s.Spec.Image == "" && len(s.Spec.Sidecars) > 0 {
		return s.Spec.Sidecars[0].Image
	}
	return s.Spec.Image
}

func getDaemonSetImage(d *appsv1.DaemonSet) string {
	if len(d.Spec.Template.Spec.Containers) > 0 {
		return d.Spec.Template.Spec.Containers[0].Image
	}
	return ""
}

func getDeploymentImage(d *appsv1.Deployment) string {
	if len(d.Spec.Template.Spec.Containers) > 0 {
		return d.Spec.Template.Spec.Containers[0].Image
	}
	return ""
}

func FormatImage(data interface{}) (string, error) {
	image := ""
	switch v := data.(type) {
	case *v1.Service:
		image = getServiceImage(v)
	case *appsv1.Deployment:
		image = getDeploymentImage(v)
	case *appsv1.DaemonSet:
		image = getDaemonSetImage(v)
	}

	return strings.TrimPrefix(image, constants.LocalRegistry+"/"), nil
}

func waitingOnBuild(svc *riov1.Service) bool {
	if svc.Spec.Image == "" && len(svc.Spec.Sidecars) == 0 {
		return true
	}

	for _, container := range svc.Spec.Sidecars {
		if container.Image == "" {
			return true
		}
	}

	return false
}
