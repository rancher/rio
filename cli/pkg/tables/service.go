package tables

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/rio/cli/pkg/table"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/services"
)

func NewService(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{.ID}"},
		{"ENDPOINT", "{{.Service.Status.Endpoints[0]}}"},
		{"SCALE", "{{scale .Service .Service.Status.ScaleStatus}}"},
		{"IMAGE", "{{.Service | image}}"},
		{"APP", "{{.Service | app}}"},
		{"VERSION", "{{.Service | version}}"},
		{"WEIGHT", "{{.Service.Spec.Weight}}"},
		{"CREATED", "{{.Service.CreationTimestamp | ago}}"},
		{"DETAIL", "{{}}"},
	}, cfg)

	writer.AddFormatFunc("image", FormatImage)
	writer.AddFormatFunc("scale", formatRevisionScale)
	writer.AddFormatFunc("app", app)
	writer.AddFormatFunc("version", version)

	return &tableWriter{
		writer: writer,
	}
}

func app(data interface{}) string {
	s, ok := data.(*v1.Service)
	if !ok {
		return ""
	}
	appName, _ := services.AppAndVersion(s)
	return appName
}

func version(data interface{}) string {
	s, ok := data.(*v1.Service)
	if !ok {
		return ""
	}
	_, version := services.AppAndVersion(s)
	return version
}

func formatRevisionScale(svc *riov1.Service, scaleStatus *v1.ScaleStatus) (string, error) {
	scale := svc.Spec.Replicas
	if svc.Status.ComputedReplicas != nil {
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
	ready := scaleNum - scaleStatus.Unavailable
	if scaleNum > 0 {
		percentage = fmt.Sprintf(" %d%%", (ready*100)/scaleNum)
	}

	if ready != scaleNum {
		prefix = fmt.Sprintf("%d/", ready)
	}

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
	return strings.TrimPrefix(image, "localhost:5442/"), nil
}
