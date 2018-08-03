package mapper

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/norman/types/mapper"
	"github.com/rancher/rio/cli/cmd/route"
)

type To struct {
	Field string
}

func (d To) FromInternal(data map[string]interface{}) {
	v, ok := data[d.Field]
	if !ok {
		return
	}

	var result []interface{}
	for _, obj := range convert.ToMapSlice(v) {
		str := destionationMapToString(obj)
		if str == "" {
			continue
		}
		result = append(result, str)
	}

	data[d.Field] = result
}

func (d To) ToInternal(data map[string]interface{}) error {
	v, ok := data[d.Field]
	if !ok {
		return nil
	}

	slc, ok := v.([]interface{})
	if !ok {
		return nil
	}

	var newSlc []interface{}

	for _, obj := range slc {
		str, ok := obj.(string)
		if ok {
			d, err := route.ParseDestinations([]string{str})
			if err != nil {
				return err
			}
			newSlc = append(newSlc, d)
		} else {
			newSlc = append(newSlc, obj)
		}
	}

	data[d.Field] = newSlc
	return nil
}

func (d To) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return mapper.ValidateField(d.Field, schema)
}
