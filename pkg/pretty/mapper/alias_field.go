package mapper

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/mapper"
)

type AliasField struct {
	Field string
	Names []string
}

func (d AliasField) FromInternal(data map[string]interface{}) {
}

func (d AliasField) ToInternal(data map[string]interface{}) error {
	for _, name := range d.Names {
		if v, ok := data[name]; ok {
			delete(data, name)
			data[d.Field] = v
		}
	}
	return nil
}

func (d AliasField) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return mapper.ValidateField(d.Field, schema)
}
