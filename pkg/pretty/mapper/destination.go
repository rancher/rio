package mapper

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/norman/types/mapper"
	"github.com/rancher/rio/cli/cmd/route"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
)

type Destination struct {
	Field string
}

func (d Destination) FromInternal(data map[string]interface{}) {
	v, ok := data[d.Field]
	if !ok {
		return
	}

	m := convert.ToMapInterface(v)
	data[d.Field] = destionationMapToString(m)
}

func destionationMapToString(m map[string]interface{}) string {
	service, _ := m["service"].(string)
	stack, _ := m["stack"].(string)
	revision, _ := m["revision"].(string)
	port, err := convert.ToNumber(m["port"])
	if err != nil {
		port = 0
	}

	weight, err := convert.ToNumber(m["weight"])
	if err != nil {
		weight = 0
	}

	wd := v1beta1.WeightedDestination{
		Weight: weight,
		Destination: v1beta1.Destination{
			Port:     port,
			Revision: revision,
			Service:  service,
			Stack:    stack,
		},
	}

	return wd.String()
}

func (d Destination) ToInternal(data map[string]interface{}) error {
	v, ok := data[d.Field]
	if !ok {
		return nil
	}

	if str, ok := v.(string); ok {
		dest, err := parseDestination(str)
		if err != nil {
			return err
		}

		data[d.Field] = dest
	}

	return nil
}

func parseDestination(str string) (map[string]interface{}, error) {
	dests, err := route.ParseDestinations([]string{str})
	if err != nil {
		return nil, err
	}

	return convert.EncodeToMap(dests[0])
}

func (d Destination) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return mapper.ValidateField(d.Field, schema)
}
