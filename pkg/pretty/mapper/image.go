package mapper

import types "github.com/rancher/mapper"

type Image struct{}

func (i Image) FromInternal(data map[string]interface{}) {
	v, ok := data["image"]
	if !ok {
		return
	}
	if _, ok = v.(string); ok {
		return
	}
	if m, ok := v.(map[string]interface{}); ok {
		data["build"] = m
	}
}

func (i Image) ToInternal(data map[string]interface{}) {

}

func (i Image) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return nil
}
