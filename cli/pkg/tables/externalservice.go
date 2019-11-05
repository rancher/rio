package tables

import (
	"fmt"
	"strings"

	"github.com/rancher/rio/cli/pkg/table"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

func NewExternalService(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{id .Obj}}"},
		{"CREATED", "{{.Obj.CreationTimestamp | ago}}"},
		{"TARGET", "{{.Obj | formatTarget}}"},
	}, cfg)
	writer.AddFormatFunc("formatTarget", FormatTarget)

	return &tableWriter{
		writer: writer,
	}
}

func FormatTarget(obj interface{}) (string, error) {
	switch item := obj.(*v1.ExternalService); {
	case item.Spec.FQDN != "":
		return item.Spec.FQDN, nil
	case item.Spec.TargetRouter != "":
		return fmt.Sprintf("%s:router/%s", item.Spec.TargetServiceNamespace, item.Spec.TargetRouter), nil
	case item.Spec.TargetApp != "" && item.Spec.TargetVersion != "":
		return fmt.Sprintf("%s:%s@%s", item.Spec.TargetServiceNamespace, item.Spec.TargetApp, item.Spec.TargetVersion), nil
	case item.Spec.TargetApp != "":
		return fmt.Sprintf("%s:%s", item.Spec.TargetServiceNamespace, item.Spec.TargetApp), nil
	case len(item.Spec.IPAddresses) > 0:
		return strings.Join(item.Spec.IPAddresses, ","), nil
	default:
		return "", nil
	}
}
