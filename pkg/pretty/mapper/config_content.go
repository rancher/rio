package mapper

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/norman/types/mapper"
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
	return mapper.ValidateField("content", schema)
}
