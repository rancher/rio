package tables

import "github.com/rancher/rio/cli/pkg/table"

func NewFeatures(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{.Name}}"},
		{"DESCRIPTION", "{{.Description}}"},
		{"ENABLED", "{{.Enabled | boolToStar}}"},
	}, cfg)

	return &tableWriter{
		writer: writer,
	}
}
