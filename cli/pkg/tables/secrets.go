package tables

import "github.com/rancher/rio/cli/pkg/table"

func NewSecret(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{id .Obj}}"},
		{"TYPE", "{{.Obj.Type}}"},
		{"DATA", "{{.Obj.Data | len}}"},
		{"CREATED", "{{.Obj.CreationTimestamp | ago}}"},
	}, cfg)

	return &tableWriter{
		writer: writer,
	}
}
