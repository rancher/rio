package tables

import (
	"encoding/base64"

	units "github.com/docker/go-units"
	"github.com/rancher/rio/cli/pkg/table"
	corev1 "k8s.io/api/core/v1"
)

func NewConfig(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{stackScopedName .Obj.Namespace .Obj.Name ``}}"},
		{"CREATED", "{{.Obj.CreationTimestamp | ago}}"},
		{"SIZE", "{{.Obj | size}}"},
	}, cfg)

	writer.AddFormatFunc("size", Base64Size)
	writer.AddFormatFunc("stackScopedName", table.FormatStackScopedName(cfg.GetSetNamespace()))

	return &tableWriter{
		writer: writer,
	}
}

func Base64Size(data interface{}) (string, error) {
	c, ok := data.(*corev1.ConfigMap)
	if !ok {
		return "", nil
	}

	size := len(c.Data) + len(c.BinaryData)
	if size > 0 {
		for _, v := range c.Data {
			size += len(v)
		}
		for _, v := range c.BinaryData {
			size += len(base64.StdEncoding.EncodeToString(v))
		}
	}

	return units.HumanSize(float64(size)), nil
}
