package tables

import (
	"github.com/rancher/rio/cli/pkg/table"
)

func NewPublicDomain(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"DOMAIN", "Obj.Spec.DomainName"},
		{"TARGET", "{{stackScopedName .Obj.Spec.TargetStackName .Obj.Spec.TargetName}}"},
	}, cfg)

	writer.AddFormatFunc("stackScopedName", table.FormatStackScopedName(cfg.GetDefaultNamespace()))
	return &tableWriter{
		writer: writer,
	}
}
