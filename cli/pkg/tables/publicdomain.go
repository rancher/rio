package tables

import (
	"github.com/rancher/rio/cli/pkg/table"
)

func NewPublicDomain(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{id .Obj}}"},
		{"DOMAIN", "Obj.Spec.DomainName"},
		{"TARGET", "{{stackScopedName .Obj.Namespace .Obj.Spec.TargetServiceName ``}}"},
	}, cfg)

	return &tableWriter{
		writer: writer,
	}
}
