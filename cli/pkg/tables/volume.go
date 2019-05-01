package tables

import (
	"github.com/rancher/mapper/convert"
	"github.com/rancher/rio/cli/pkg/table"
)

type Config interface {
	table.WriterConfig
	GetDefaultStackName() string
	Domain() (string, error)
}

func NewVolume(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{stackScopedName .Obj.Namespace .Obj.ServiceName}}"},
		{"DRIVER", "{{.Obj.Spec.Driver | driver}}"},
		{"TEMPLATE", "Obj.Spec.Template"},
		{"SIZE GB", "Obj.Spec.SizeInGB"},
		{"CREATED", "{{.Obj.CreationTimestamp | ago}}"},
	}, cfg)
	defer writer.Close()

	writer.AddFormatFunc("driver", FormatDriver)
	writer.AddFormatFunc("stackScopedName", table.FormatStackScopedName(cfg.GetDefaultStackName()))

	return &tableWriter{
		writer: writer,
	}
}

func FormatDriver(obj interface{}) (string, error) {
	str := convert.ToString(obj)
	if str == "" {
		return "default", nil
	}
	return str, nil
}
