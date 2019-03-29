package tables

import (
	"github.com/rancher/rio/cli/pkg/table"
)

func NewFeature(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "Obj.Name"},
		{"ENABLED", "{{.Obj.Spec.Enabled | boolToStar}}"},
		{"DESCRIPTION", "Obj.Spec.Description"},
	}, cfg)

	return &tableWriter{
		writer: writer,
	}

}
