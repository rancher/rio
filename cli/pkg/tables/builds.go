package tables

import (
	"fmt"

	"github.com/rancher/rio/cli/pkg/table"
)

func NewBuild(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{stackScopedName .Obj.Namespace .Obj.Name ``}}"},
		{"CREATED", "{{.Obj.CreationTimestamp | ago}}"},
		{"SERVICE", "{{.Obj.Labels | findService}}"},
	}, cfg)

	writer.AddFormatFunc("stackScopedName", table.FormatStackScopedName(cfg.GetSetNamespace()))
	writer.AddFormatFunc("findService", findService)
	return &tableWriter{
		writer: writer,
	}
}

func findService(data interface{}) (string, error) {
	m, ok := data.(map[string]string)
	if !ok {
		return "", nil
	}
	name := m["service-name"]
	namespace := m["service-namespace"]
	return fmt.Sprintf("%s/%s", namespace, name), nil
}
