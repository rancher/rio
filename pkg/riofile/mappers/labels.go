package mappers

import (
	"strings"

	"github.com/rancher/norman/pkg/types/mapper"

	"github.com/rancher/norman/pkg/data"
	"github.com/rancher/norman/pkg/types"
)

type LabelCleaner struct {
	mapper.DefaultMapper
}

func (d LabelCleaner) FromInternal(data data.Object) {
	annotations, ok := data["annotations"].(map[string]interface{})
	if ok {
		for k := range annotations {
			if strings.Contains(k, "rio.cattle.io") {
				delete(annotations, k)
			}
		}
		if len(annotations) == 0 {
			delete(data, "annotations")
		}
	}

	labels, ok := data["labels"].(map[string]interface{})
	if ok {
		for k := range labels {
			if strings.Contains(k, "rio.cattle.io") {
				delete(labels, k)
			}
		}
		if len(labels) == 0 {
			delete(data, "labels")
		}
	}
}

func (d LabelCleaner) ToInternal(data data.Object) error {
	return nil
}

func (d LabelCleaner) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return nil
}
