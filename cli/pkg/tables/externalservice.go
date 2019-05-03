package tables

import (
	"strings"

	"github.com/rancher/rio/cli/pkg/table"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

func NewExternalService(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{stackScopedName .Obj.SystemNamespace .Obj.ServiceName}}"},
		{"CREATED", "{{.Obj.Created | ago}}"},
		{"TARGET", "{{.Obj | formatTarget}}"},
	}, cfg)
	writer.AddFormatFunc("stackScopedName", table.FormatStackScopedName(cfg.GetDefaultNamespace()))
	writer.AddFormatFunc("formatTarget", FormatTarget)

	return &tableWriter{
		writer: writer,
	}
}

func FormatTarget(obj interface{}) (string, error) {
	item := obj.(*v1.ExternalService)
	endpoint := ""
	if item.Spec.FQDN != "" {
		endpoint = item.Spec.FQDN
	} else if item.Spec.Service != "" {
		endpoint = item.Spec.Service
	} else if len(item.Spec.IPAddresses) > 0 {
		endpoint = strings.Join(item.Spec.IPAddresses, ",")
	}

	return endpoint, nil
}
