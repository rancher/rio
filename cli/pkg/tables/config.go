package tables

import (
	"encoding/base64"

	"github.com/docker/go-units"
	"github.com/rancher/rio/cli/pkg/table"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

func NewConfig(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{stackScopedName .Obj.Namespace .Obj.Name}}"},
		{"CREATED", "{{.Obj.CreationTimestamp | ago}}"},
		{"SIZE", "{{.Obj | size}}"},
	}, cfg)

	writer.AddFormatFunc("size", Base64Size)
	writer.AddFormatFunc("stackScopedName", table.FormatStackScopedName(cfg.GetDefaultStackName()))

	return &tableWriter{
		writer: writer,
	}
}

func Base64Size(data interface{}) (string, error) {
	c, ok := data.(*v1.Config)
	if !ok {
		return "", nil
	}

	size := len(c.Spec.Content)
	if size > 0 && c.Spec.Encoded {
		content, err := base64.StdEncoding.DecodeString(c.Spec.Content)
		if err != nil {
			return "", err
		}
		size = len(content)
	}

	return units.HumanSize(float64(size)), nil
}
