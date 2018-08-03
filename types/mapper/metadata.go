package mapper

import (
	"encoding/json"

	"strings"

	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/norman/types/mapper"
)

func NewMetadata(field string) *Metadata {
	return &Metadata{Field: field}
}

type Metadata struct {
	Field string
}

func (d Metadata) FromInternal(data map[string]interface{}) {
	data = convert.ToMapInterface(data[d.Field])
	for key, value := range data {
		str := convert.ToString(value)
		if strings.HasPrefix(str, "{") {
			newValue := map[string]interface{}{}
			if err := json.Unmarshal([]byte(str), &newValue); err != nil {
				data[key] = newValue
			}
		} else if strings.HasPrefix(str, "[") {
			var newValue []interface{}
			if err := json.Unmarshal([]byte(str), &newValue); err != nil {
				data[key] = newValue
			}
		}
	}
}

func (d Metadata) ToInternal(data map[string]interface{}) error {
	data = convert.ToMapInterface(data[d.Field])
	for key, value := range data {
		switch value.(type) {
		case []interface{}:
			bytes, err := json.Marshal(value)
			if err != nil {
				return err
			}
			data[key] = string(bytes)
		case []map[string]interface{}:
			bytes, err := json.Marshal(value)
			if err != nil {
				return err
			}
			data[key] = string(bytes)
		default:
			data[key] = convert.ToString(value)
		}
	}

	return nil
}

func (d Metadata) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	if err := mapper.ValidateField(d.Field, schema); err != nil {
		return err
	}
	f := schema.ResourceFields[d.Field]
	f.Type = "map[json]"
	schema.ResourceFields[d.Field] = f
	return nil
}
