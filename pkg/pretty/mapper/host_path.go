package mapper

import (
	"strings"

	types "github.com/rancher/mapper"
	"github.com/rancher/mapper/convert"
	"github.com/rancher/mapper/mappers"
)

type HostPath struct {
	Field string
}

func (d HostPath) FromInternal(data map[string]interface{}) {
	v, ok := data[d.Field]
	if !ok {
		return
	}

	m := convert.ToMapInterface(v)

	host, _ := m["host"].(string)
	path, _ := m["path"].(string)

	data[d.Field] = host + path
}

func (d HostPath) ToInternal(data map[string]interface{}) error {
	v, ok := data[d.Field]
	if !ok {
		return nil
	}

	if m, ok := v.(string); ok {
		parts := strings.SplitAfterN(m, "/", 2)
		if len(parts) == 2 {
			data[d.Field] = map[string]interface{}{
				"host": parts[0],
				"path": parts[1],
			}
		} else if parts[0][0] == '/' {
			data[d.Field] = map[string]interface{}{
				"path": parts[0],
			}
		} else {
			data[d.Field] = map[string]interface{}{
				"host": parts[0],
			}
		}
	}

	return nil
}

func (d HostPath) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return mappers.ValidateField(d.Field, schema)
}
