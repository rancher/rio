package mapper

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/norman/types/mapper"
)

type SingleSlice struct {
	Field string
}

func (d SingleSlice) FromInternal(data map[string]interface{}) {
	v, ok := data[d.Field]
	if !ok {
		return
	}

	ss := convert.ToInterfaceSlice(v)
	if len(ss) == 1 {
		if _, ok := ss[0].(string); ok {
			data[d.Field] = ss[0]
		}
	}
}

func (d SingleSlice) ToInternal(data map[string]interface{}) error {
	v, ok := data[d.Field]
	if !ok {
		return nil
	}

	if str, ok := v.(string); ok {
		data[d.Field] = []interface{}{str}
	}

	return nil
}

func (d SingleSlice) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return mapper.ValidateField(d.Field, schema)
}
