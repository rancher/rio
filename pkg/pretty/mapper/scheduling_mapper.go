package mapper

import (
	types "github.com/rancher/mapper"
	"github.com/rancher/mapper/convert"
	"github.com/rancher/mapper/mappers"
	"github.com/rancher/mapper/values"
)

type SchedulingMapper struct {
	Field string
}

func (d SchedulingMapper) FromInternal(data map[string]interface{}) {
	g := values.GetValueN

	scheduling := convert.ToMapInterface(g(data, d.Field))
	if len(scheduling) != 1 {
		return
	}

	requireAll := convert.ToMapInterface(g(data, d.Field, "node", "requireAll"))
	requireAny := convert.ToMapInterface(g(data, d.Field, "node", "requireAny"))
	preferred := convert.ToMapInterface(g(data, d.Field, "node", "preferred"))

	if len(preferred) == 0 && len(requireAny) == 0 && len(requireAll) > 0 {
		data[d.Field] = requireAll
	}
}

func (d SchedulingMapper) ToInternal(data map[string]interface{}) error {
	requiredAllSlice := convert.ToInterfaceSlice(data[d.Field])
	if len(requiredAllSlice) > 0 {
		data[d.Field] = map[string]interface{}{
			"node": map[string]interface{}{
				"requireAll": requiredAllSlice,
			},
		}
	}

	return nil
}

func (d SchedulingMapper) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return mappers.ValidateField(d.Field, schema)
}
