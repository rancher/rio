package mappers

import (
	"fmt"
	"strings"

	"github.com/rancher/mapper/convert"
	"github.com/rancher/mapper/mappers"
	"github.com/rancher/wrangler/pkg/kv"
)

type Fuzzy struct {
	mappers.DefaultMapper
	field string
	vals  map[string]string
}

func NewFuzzy(field string, vals ...string) Fuzzy {
	f := Fuzzy{
		DefaultMapper: mappers.DefaultMapper{
			Field: field,
		},
		vals: map[string]string{},
	}

	for _, v := range vals {
		k := v
		if strings.Contains(v, "=") {
			v, k = kv.Split(v, "=")
		}
		f.vals[convert.ToYAMLKey(v)] = k
		f.vals[strings.Replace(convert.ToYAMLKey(v), "_", "-", -1)] = k
		f.vals[strings.ToLower(v)] = k
		f.vals[v] = k
	}

	return f
}

func (d Fuzzy) FromInternal(data map[string]interface{}) {
}

func (d Fuzzy) ToInternal(data map[string]interface{}) error {
	v, ok := data[d.Field]
	if ok {
		newValue, ok := d.vals[convert.ToString(v)]
		if !ok {
			return fmt.Errorf("%s is not a valid value for field %s", v, d.Field)
		}
		data[d.Field] = newValue
	}
	return nil
}
