package mappers

import (
	"strings"

	"github.com/rancher/mapper"
	"github.com/rancher/mapper/mappers"
)

type LabelCleaner struct {
	mappers.DefaultMapper
}

func (d LabelCleaner) FromInternal(data map[string]interface{}) {
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

func (d LabelCleaner) ToInternal(data map[string]interface{}) error {
	return nil
}

func (d LabelCleaner) ModifySchema(schema *mapper.Schema, schemas *mapper.Schemas) error {
	return nil
}
