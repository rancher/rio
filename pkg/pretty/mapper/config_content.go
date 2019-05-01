package mapper

import (
	types "github.com/rancher/mapper"
	"github.com/rancher/mapper/convert"
	"github.com/rancher/mapper/mappers"
)

type ConfigContent struct {
}

func (d ConfigContent) FromInternal(data map[string]interface{}) {
	content, contentOk := data["content"]
	encoded, ok := data["encoded"]
	if ok && contentOk && convert.ToBool(encoded) {
		delete(data, "content")
		data["encoded"] = content
	}
}

func (d ConfigContent) ToInternal(data map[string]interface{}) error {
	_, ok := data["content"]
	if ok {
		data["encoded"] = false
	}
	encoded, ok := data["encoded"]
	_, isStr := encoded.(string)
	if ok && isStr {
		data["content"] = encoded
		data["encoded"] = true
	}
	return nil
}

func (d ConfigContent) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return mappers.ValidateField("content", schema)
}
