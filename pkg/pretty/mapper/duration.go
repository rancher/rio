package mapper

import (
	"time"

	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/norman/types/mapper"
	"github.com/rancher/rio/cli/cmd/create"
)

type Duration struct {
	Field string
	Unit  time.Duration
}

func (d Duration) FromInternal(data map[string]interface{}) {
	v, ok := data[d.Field]
	if !ok {
		return
	}

	n, err := convert.ToNumber(v)
	if err != nil {
		return
	}

	data[d.Field] = (time.Duration(n) * d.Unit).String()
}

func (d Duration) ToInternal(data map[string]interface{}) error {
	v, ok := data[d.Field]
	if !ok {
		return nil
	}

	if str, ok := v.(string); ok {
		sec, err := create.ParseDurationUnit(str, d.Field, d.Unit)
		if err != nil {
			return err
		}
		data[d.Field] = sec
	}

	return nil
}

func (d Duration) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return mapper.ValidateField(d.Field, schema)
}
