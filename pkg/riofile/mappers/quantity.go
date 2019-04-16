package mappers

import (
	"github.com/rancher/mapper/mappers"
	"github.com/rancher/rio/pkg/pretty/stringers"
	"k8s.io/apimachinery/pkg/api/resource"
)

type QuantityMapper struct {
	mappers.DefaultMapper
}

func NewQuantity(field string) QuantityMapper {
	return QuantityMapper{
		DefaultMapper: mappers.DefaultMapper{
			Field: field,
		},
	}
}

func (d QuantityMapper) FromInternal(data map[string]interface{}) {
	v, ok := data[d.Field]
	if !ok {
		return
	}

	if q, ok := v.(resource.Quantity); ok {
		data[d.Field] = q.String()
	}
}

func (d QuantityMapper) ToInternal(data map[string]interface{}) error {
	v, ok := data[d.Field]
	if !ok {
		return nil
	}

	if s, ok := v.(string); ok {
		q, err := stringers.ParseQuantity(s)
		if err != nil {
			return err
		}
		if !q.IsZero() {
			data[d.Field] = q
		}
	}

	return nil
}
