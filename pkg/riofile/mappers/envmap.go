package mappers

import (
	"fmt"
	"strings"

	"github.com/rancher/wrangler/pkg/data"
	types "github.com/rancher/wrangler/pkg/schemas"
	mapper "github.com/rancher/wrangler/pkg/schemas/mappers"
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

	for _, opt := range opts {
		if strings.HasPrefix(opt, "sep=") {
			e.Sep = strings.TrimPrefix(opt, "sep=")
		}
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
