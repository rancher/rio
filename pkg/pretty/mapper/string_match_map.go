package mapper

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/norman/types/mapper"
)

type StringMatchMap struct {
	Field string
}

func (d StringMatchMap) FromInternal(data map[string]interface{}) {
	v, ok := data[d.Field]
	if !ok {
		return
	}

	m := convert.ToMapInterface(v)

	for k, v := range m {
		m[k] = stringMatchToString(v)
	}
}

func (d StringMatchMap) ToInternal(data map[string]interface{}) error {
	v, ok := data[d.Field]
	if !ok {
		return nil
	}

	m := convert.ToMapInterface(v)
	for k, v := range m {
		if str, ok := v.(string); ok {
			m[k] = ParseStringMatch(str)
		}
	}

	return nil
}

func (d StringMatchMap) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return mapper.ValidateField(d.Field, schema)
}
