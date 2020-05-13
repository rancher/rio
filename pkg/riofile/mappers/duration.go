package mappers

import (
	"github.com/rancher/rio/pkg/riofile/stringers"
	"github.com/rancher/wrangler/pkg/data"
	types "github.com/rancher/wrangler/pkg/schemas"
	mapper "github.com/rancher/wrangler/pkg/schemas/mappers"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DurationMapper struct {
	mapper.DefaultMapper
}

func NewDuration(field string, args ...string) types.Mapper {
	return DurationMapper{
		DefaultMapper: mapper.DefaultMapper{
			Field: field,
		},
	}
}

func (d DurationMapper) FromInternal(data data.Object) {
	v, ok := data[d.Field]
	if !ok {
		return
	}

	if q, ok := v.(v1.Duration); ok {
		data[d.Field] = q.String()
	}
}

func (d DurationMapper) ToInternal(data data.Object) error {
	v, ok := data[d.Field]
	if !ok {
		return nil
	}

	if s, ok := v.(string); ok {
		q, err := stringers.ParseDuration(s)
		if err != nil {
			return err
		}
		data[d.Field] = q
	}

	return nil
}
