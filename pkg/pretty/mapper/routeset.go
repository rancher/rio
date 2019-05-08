package mapper

import (
	types "github.com/rancher/mapper"
	"github.com/rancher/mapper/convert"
	"github.com/rancher/mapper/mappers"
)

type RouteSet struct {
	Field string
}

func (r RouteSet) FromInternal(data map[string]interface{}) {
	d, ok := data[r.Field]
	if !ok {
		return
	}

	dm := convert.ToMapInterface(d)
	for k, v := range dm {
		vm := convert.ToMapInterface(v)
		routes, ok := vm["routes"]
		if ok {
			dm[k] = routes
		}
	}
}

func (r RouteSet) ToInternal(data map[string]interface{}) error {
	d, ok := data[r.Field]
	if !ok {
		return nil
	}

	dm := convert.ToMapInterface(d)
	for k, v := range dm {
		if sl, ok := v.([]interface{}); ok {
			dm[k] = map[string]interface{}{
				"routes": sl,
			}
		}
	}

	return nil
}

func (r RouteSet) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return mappers.ValidateField(r.Field, schema)
}
