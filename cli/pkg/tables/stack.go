package tables

import (
	"github.com/rancher/rio/cli/pkg/table"
)

func NewStack(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "Obj.ServiceName"},
		{"STATE", "Obj | toJson | state"},
		{"CREATED", "{{.Obj.CreationTimestamp | ago}}"},
		{"DESC", "Obj.Spec.Description"},
		{"DETAIL", "Obj | transitioning"},
	}, cfg)

	return &tableWriter{
		writer: writer,
	}
}
