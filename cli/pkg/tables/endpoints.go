package tables

import "github.com/rancher/rio/cli/pkg/table"

func NewEndpoint(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{.Name}}"},
		{"ENDPOINTS", "{{.Endpoints | array}}"},
	}, cfg)

	writer.AddFormatFunc("size", Base64Size)

	return &tableWriter{
		writer: writer,
	}
}
