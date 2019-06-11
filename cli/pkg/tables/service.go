package tables

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/rio/cli/pkg/table"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/services"
	corev1 "k8s.io/api/core/v1"
)

func NewService(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"Name", "{{serviceName .Service.Namespace .Service}}"},
		{"IMAGE", "{{.Service | image}}"},
		{"CREATED", "{{.Service.CreationTimestamp | ago}}"},
		{"SCALE", "{{scale .Service.Spec.Scale .Service.Status.ScaleStatus .Service.Status.ObservedScale}}"},
		{"ENDPOINT", "{{.Service.Status.Endpoints | array}}"},
		{"WEIGHT", "{{.Service.Spec.Weight}}"},
		{"DETAIL", "{{.Pods | podsDetail}}"},
	}, cfg)

	writer.AddFormatFunc("serviceName", FormatServiceName(cfg))
	writer.AddFormatFunc("image", FormatImage)
	writer.AddFormatFunc("scale", FormatScale)
	writer.AddFormatFunc("podsDetail", podsDetail)

	return &tableWriter{
		writer: writer,
	}
}

func podsDetail(obj interface{}) (string, error) {
	pods, _ := obj.([]corev1.Pod)

	if len(pods) == 0 {
		return "", nil
	}
	return podDetail(&pods[0])
}

func FormatScale(data, data2, data3 interface{}) (string, error) {
	scale, ok := data.(int)
	if !ok {
		return fmt.Sprint(data), nil
	}
	if scale == 0 {
		scale = 1
	}
	observedScale, ok := data3.(*int)
	if ok && observedScale != nil {
		scale = *observedScale
	}
	scaleStr := strconv.Itoa(scale)

	scaleStatus, ok := data2.(*v1.ScaleStatus)
	if !ok {
		return scaleStr, nil
	}

	if scaleStatus == nil {
		scaleStatus = &v1.ScaleStatus{}
	}

	if scaleStatus.Available == 0 && scaleStatus.Unavailable == 0 && scaleStatus.Ready == scale {
		return scaleStr, nil
	}

	percentage := ""
	if scale > 0 && scaleStatus.Updated > 0 && scale != scaleStatus.Updated {
		percentage = fmt.Sprintf(" %d%%", (scaleStatus.Updated*100)/scale)
	}

	prefix := ""
	if scale > 0 && scaleStatus.Ready != scale {
		prefix = fmt.Sprintf("%d/", scaleStatus.Ready)
	}

	//return fmt.Sprintf("(%d/%d/%d)/%d%s", scaleStatus.Unavailable, scaleStatus.Available, scaleStatus.Ready, scale, percentage), nil
	return fmt.Sprintf("%s%d%s", prefix, scale, percentage), nil
}

func FormatServiceName(cfg Config) func(data, data2 interface{}) (string, error) {
	return func(data, data2 interface{}) (string, error) {
		ns, ok := data.(string)
		if !ok {
			return "", nil
		}

		service, ok := data2.(*v1.Service)
		if !ok {
			return "", nil
		}

		app, version := services.AppAndVersion(service)

		return table.FormatStackScopedName(cfg.GetSetNamespace())(ns, app, version)
	}
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
	return strings.TrimPrefix(image, "localhost:5442/"), nil
}
