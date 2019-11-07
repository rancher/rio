package tables

import (
	"fmt"

	"github.com/rancher/rio/cli/pkg/table"
)

func NewEndpoint(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{endpointName .Name .Namespace}}"},
		{"ENDPOINTS", "{{.Endpoints | array}}"},
	}, cfg)

	writer.AddFormatFunc("endpointName", func(name, namespace string) string {
		if cfg.GetSetNamespace() == "" {
			return fmt.Sprintf("%s:%s", namespace, name)
		}
		return name
	})
	writer.AddFormatFunc("size", Base64Size)

	return &tableWriter{
		writer: writer,
	}
}
