package tables

import (
	"github.com/rancher/rio/cli/pkg/table"
)

func NewPublicDomain(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"DOMAIN", "Obj.Spec.DomainName"},
		{"TARGET", "{{stackScopedName .Obj.Namespace .Obj.Spec.TargetServiceName ``}}"},
	}, cfg)

	writer.AddFormatFunc("stackScopedName", table.FormatStackScopedName(cfg.GetSetNamespace()))
	return &tableWriter{
		writer: writer,
	}
}
