package mappers

import (
	"fmt"

	"github.com/rancher/norman/pkg/data"
	"github.com/rancher/norman/pkg/types"
	"github.com/rancher/norman/pkg/types/mapper"
)

type EnvMapper struct {
	mapper.DefaultMapper
	Sep string
}

func NewEnvMap(field string, opts ...string) types.Mapper {
	e := EnvMapper{
		DefaultMapper: mapper.DefaultMapper{
			Field: field,
		},
	}

	if len(opts) > 0 {
		e.Sep = opts[0]
	}

	return e
}

func (e EnvMapper) ToInternal(data data.Object) error {
	m := data.Map(e.Field)
	if m == nil {
		return nil
	}

	var result []interface{}
	for k, v := range m {
		item := fmt.Sprintf("%s%s%s", k, e.Sep, v)
		result = append(result, item)
	}

	data.Set(e.Field, result)
	return nil
}
