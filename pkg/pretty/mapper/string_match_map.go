package mapper

import (
	types "github.com/rancher/mapper"
	"github.com/rancher/mapper/convert"
	"github.com/rancher/mapper/mappers"
	"github.com/rancher/rio/pkg/pretty/objectmappers"
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

	for k := range m {
		objectmappers.NewStringMatch(k).FromInternal(m)
	}
}

func (d StringMatchMap) ToInternal(data map[string]interface{}) error {
	v, ok := data[d.Field]
	if !ok {
		return nil
	}

	m := convert.ToMapInterface(v)
	for k := range m {
		if err := objectmappers.NewStringMatch(k).ToInternal(m); err != nil {
			return err
		}
	}

	return nil
}

func (d StringMatchMap) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return mappers.ValidateField(d.Field, schema)
}
