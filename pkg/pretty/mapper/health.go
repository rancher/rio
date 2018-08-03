package mapper

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/norman/types/mapper"
)

type HealthMapper struct {
	Field string
}

func (d HealthMapper) FromInternal(data map[string]interface{}) {
	v, ok := data[d.Field]
	if !ok {
		return
	}

	m := convert.ToMapInterface(v)
	if len(m) == 1 {
		if len(convert.ToInterfaceSlice(m["test"])) == 1 {
			data[d.Field] = convert.ToInterfaceSlice(m["test"])[0]
		} else if str, ok := m["test"].(string); ok {
			data[d.Field] = str
		}
	}
}

func (d HealthMapper) ToInternal(data map[string]interface{}) error {
	v, ok := data[d.Field]
	if !ok {
		return nil
	}

	if str, ok := v.(string); ok {
		data[d.Field] = map[string]interface{}{
			"test": str,
		}
	}

	return nil
}

func (d HealthMapper) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return mapper.ValidateField(d.Field, schema)
}
