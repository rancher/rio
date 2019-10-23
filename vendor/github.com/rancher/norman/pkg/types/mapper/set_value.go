package mapper

import (
	"github.com/rancher/norman/pkg/data"
	"github.com/rancher/norman/pkg/types"
)

type SetValue struct {
	Field         string
	InternalValue interface{}
	ExternalValue interface{}
}

func (d SetValue) FromInternal(data data.Object) {
	if d.ExternalValue != nil {
		data[d.Field] = d.ExternalValue
	}
}

func (d SetValue) ToInternal(data data.Object) error {
	if d.InternalValue != nil {
		data[d.Field] = d.InternalValue
	}
	return nil
}

func (d SetValue) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return ValidateField(d.Field, schema)
}
