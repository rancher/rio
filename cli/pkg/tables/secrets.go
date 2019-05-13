package tables

import "github.com/rancher/rio/cli/pkg/table"

func NewSecret(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{stackScopedName .Obj.Namespace .Obj.Name ``}}"},
		{"TYPE", "{{.Obj.Type}}"},
		{"DATA", "{{.Obj.Data | len}}"},
		{"CREATED", "{{.Obj.CreationTimestamp | ago}}"},
	}, cfg)

	writer.AddFormatFunc("stackScopedName", table.FormatStackScopedName(cfg.GetDefaultNamespace()))
	return &tableWriter{
		writer: writer,
	}
}
