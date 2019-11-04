package tables

import (
	"fmt"
	"strconv"
	"strings"

	webhookv1 "github.com/rancher/gitwatcher/pkg/apis/gitwatcher.cattle.io/v1"
	"github.com/rancher/rio/cli/pkg/table"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/controllers/pkg"
	"github.com/rancher/rio/pkg/services"
	corev1 "k8s.io/api/core/v1"
)

func NewService(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{.Name}}"},
		{"IMAGE", "{{.Service | image}}"},
		{"ENDPOINT", "{{arrayFirst .Service.Status.Endpoints}}"},
		{"SCALE", "{{scale .Service .Service.Status.ScaleStatus}}"},
		{"APP/VERSION", "{{.Service | appAndVersion}}"},
		{"WEIGHT", "{{.Service | formatWeight}}"},
		{"CREATED", "{{.Service.CreationTimestamp | ago}}"},
		{"DETAIL", "{{serviceDetail .Service .Pods .GitWatcher}}"},
	}, cfg)

	writer.AddFormatFunc("image", FormatImage)
	writer.AddFormatFunc("scale", formatRevisionScale)
	writer.AddFormatFunc("appAndVersion", appAndVersion)
	writer.AddFormatFunc("formatWeight", formatWeight)
	writer.AddFormatFunc("serviceDetail", serviceDetail)

	return &tableWriter{
		writer: writer,
	}
}

func appAndVersion(data interface{}) string {
	s, ok := data.(*v1.Service)
	if !ok {
		return ""
	}
	appName, version := services.AppAndVersion(s)
	return fmt.Sprintf("%s/%s", appName, version)
}

func formatWeight(data interface{}) string {
	s, ok := data.(*v1.Service)
	if !ok {
		return ""
	}
	if s.Status.ComputedWeight != nil {
		return fmt.Sprintf("%s%%", strconv.Itoa(*s.Status.ComputedWeight))
	}

	return "0%"
}

func serviceDetail(data interface{}, pods []*corev1.Pod, gitwatcher *webhookv1.GitWatcher) string {
	s, ok := data.(*v1.Service)
	if !ok {
		return ""
	}

	if s.Spec.Template {
		return "Build Template"
	}

	buffer := strings.Builder{}

	pd := ""
	for _, pod := range pods {
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

func formatRevisionScale(svc *riov1.Service, scaleStatus *v1.ScaleStatus) (string, error) {
	scale := svc.Spec.Replicas
	if svc.Status.ComputedReplicas != nil && services.AutoscaleEnable(svc) {
		scale = svc.Status.ComputedReplicas
	}
	return FormatScale(scale, scaleStatus)
}

func FormatScale(scale *int, scaleStatus *v1.ScaleStatus) (string, error) {
	scaleNum := 1
	if scale != nil {
		scaleNum = *scale
	}

	scaleStr := strconv.Itoa(scaleNum)

	if scaleStatus == nil {
		scaleStatus = &v1.ScaleStatus{}
	}

	if scaleNum == -1 {
		return strconv.Itoa(scaleStatus.Available), nil
	}

	if scaleStatus.Unavailable == 0 {
		return scaleStr, nil
	}

	var prefix string
	percentage := ""
	ready := scaleStatus.Available
	if scaleNum > 0 {
		percentage = fmt.Sprintf(" %d%%", (ready*100)/scaleNum)
	}

	prefix = fmt.Sprintf("%d/", ready)

	return fmt.Sprintf("%s%d%s", prefix, scaleNum, percentage), nil
}

func FormatImage(data interface{}) (string, error) {
	s, ok := data.(*v1.Service)
	if !ok {
		return fmt.Sprint(data), nil
	}
	image := ""
	if s.Spec.Image == "" && len(s.Spec.Sidecars) > 0 {
		image = s.Spec.Sidecars[0].Image
	} else {
		image = s.Spec.Image
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
