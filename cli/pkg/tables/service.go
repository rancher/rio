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

func FormatScale(data *int, data2, data3 interface{}) (string, error) {
	scale := data
	if data != nil && *data > 0 {
		return fmt.Sprint(*data), nil
	}

	scaleNum := 0
	if scale == nil {
		scaleNum = 1
	} else {
		scaleNum = *scale
	}

	observedScale, ok := data3.(*int)
	if ok && observedScale != nil {
		scaleNum = *observedScale
	}
	scaleStr := strconv.Itoa(scaleNum)

	scaleStatus, ok := data2.(*v1.ScaleStatus)
	if !ok {
		return scaleStr, nil
	}

	if scaleStatus == nil {
		scaleStatus = &v1.ScaleStatus{}
	}

	if scaleNum == -1 {
		return strconv.Itoa(scaleStatus.Ready), nil
	}

	if scaleStatus.Available == 0 && scaleStatus.Unavailable == 0 && scaleStatus.Ready == scaleNum {
		return scaleStr, nil
	}

	percentage := ""
	if scaleNum > 0 && scaleStatus.Updated > 0 && scaleNum != scaleStatus.Updated {
		percentage = fmt.Sprintf(" %d%%", (scaleStatus.Updated*100)/scaleNum)
	}

	prefix := ""
	if scaleNum > 0 && scaleStatus.Ready != scaleNum {
		prefix = fmt.Sprintf("%d/", scaleStatus.Ready)
	}

	return fmt.Sprintf("%s%d%s", prefix, scaleNum, percentage), nil
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
