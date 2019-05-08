package mappers

import (
	"github.com/rancher/mapper/convert"
	"github.com/rancher/mapper/mappers"
)

type ConfigMapMapper struct {
	mappers.DefaultMapper
}

func NewConfigMapMapper(field string) ConfigMapMapper {
	return ConfigMapMapper{
		DefaultMapper: mappers.DefaultMapper{
			Field: "data",
		},
	}
}

func (d ConfigMapMapper) FromInternal(data map[string]interface{}) {
	newData, ok := data["data"]
	if !ok {
		return
	}

	delete(data, "data")
	for k, v := range convert.ToMapInterface(newData) {
		data[k] = v
	}
}

func (d ConfigMapMapper) ToInternal(data map[string]interface{}) error {
	newData := map[string]interface{}{}
	for k, v := range data {
		if k != "labels" && v != "annotations" {
			delete(data, k)
			newData[k] = v
		}
	}

	if len(newData) > 0 {
		data["data"] = newData
	}

	return nil
}
