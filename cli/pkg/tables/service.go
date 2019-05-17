package tables

import (
	"fmt"
	"strconv"

	"github.com/rancher/rio/pkg/services"

	"github.com/rancher/rio/cli/pkg/table"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

func NewService(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{serviceName .Obj.Namespace .Obj}}"},
		{"IMAGE", "{{.Obj | image}}"},
		{"CREATED", "{{.Obj.CreationTimestamp | ago}}"},
		{"STATE", "{{.Obj | state}}"},
		{"SCALE", "{{scale .Obj.Spec.Scale .Obj.Status.ScaleStatus .Obj.Status.ObservedScale}}"},
		{"ENDPOINT", "{{.Obj.Status.Endpoints | array}}"},
		{"WEIGHT", "{{.Obj.Spec.Weight | graph}}"},
		{"DETAIL", "{{first (.Obj |  transitioning) (.Obj | transitioning)}}"},
	}, cfg)

	writer.AddFormatFunc("serviceName", FormatServiceName(cfg))
	writer.AddFormatFunc("image", FormatImage)
	writer.AddFormatFunc("scale", FormatScale)

	return &tableWriter{
		writer: writer,
	}
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
	if !ok || scaleStatus == nil {
		return scaleStr, nil
	}

	if scaleStatus.Available == 0 && scaleStatus.Unavailable == 0 && scaleStatus.Ready == scale {
		return scaleStr, nil
	}

	percentage := ""
	if scale > 0 && scaleStatus.Updated > 0 && scale != scaleStatus.Updated {
		percentage = fmt.Sprintf(" %d%%", (scaleStatus.Updated*100)/scale)
	}

	return fmt.Sprintf("(%d/%d/%d)/%d%s", scaleStatus.Unavailable, scaleStatus.Available, scaleStatus.Ready, scale, percentage), nil
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
	if s.Spec.Image == "" && len(s.Spec.Sidecars) > 0 {
		return s.Spec.Sidecars[0].Image, nil
	}
	return s.Spec.Image, nil
}
