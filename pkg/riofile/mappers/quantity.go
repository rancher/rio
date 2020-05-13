package mappers

import (
	"github.com/rancher/rio/pkg/riofile/stringers"
	"github.com/rancher/wrangler/pkg/data"
	types "github.com/rancher/wrangler/pkg/schemas"
	mapper "github.com/rancher/wrangler/pkg/schemas/mappers"
)

type QuantityMapper struct {
	mapper.DefaultMapper
}

func NewQuantity(field string, args ...string) types.Mapper {
	return QuantityMapper{
		DefaultMapper: mapper.DefaultMapper{
			Field: field,
		},
	}
}

func (d QuantityMapper) ToInternal(data data.Object) error {
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
			data[d.Field], _ = q.AsInt64()
		}
	}

	return nil
}
