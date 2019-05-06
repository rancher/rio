package tables

import (
	"github.com/rancher/rio/cli/pkg/table"
)

func NewFeature(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "Obj.Name"},
		{"ENABLED", "{{featureEnabled .Obj.Status.EnableOverride .Obj.Spec.Enabled}}"},
		{"DESCRIPTION", "Obj.Spec.Description"},
	}, cfg)

	writer.AddFormatFunc("featureEnabled", featureEnabled)

	return &tableWriter{
		writer: writer,
	}

}

func featureEnabled(data, data2 interface{}) (string, error) {
	if v, ok := data.(*bool); ok && v != nil {
		if *v {
			return "*", nil
		}
		return "", nil
	}

	if v, ok := data2.(bool); ok && v {
		return "*", nil
	}

	return "", nil
}
