package mapper

import (
	"strings"

	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/norman/types/mapper"
)

type StringMatch struct {
	Field string
}

func (d StringMatch) FromInternal(data map[string]interface{}) {
	v, ok := data[d.Field]
	if !ok {
		return
	}

	data[d.Field] = stringMatchToString(v)
}

func (d StringMatch) ToInternal(data map[string]interface{}) error {
	v, ok := data[d.Field]
	if !ok {
		return nil
	}

	if str, ok := v.(string); ok {
		data[d.Field] = ParseStringMatch(str)
	}

	return nil
}

func (d StringMatch) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return mapper.ValidateField(d.Field, schema)
}

func ParseStringMatch(str string) map[string]interface{} {
	if strings.HasSuffix(str, "*") {
		return map[string]interface{}{
			"prefix": str[:len(str)-1],
		}
	} else if (strings.HasPrefix(str, "regex(") || strings.HasPrefix(str, "regexp(")) && strings.HasSuffix(str, ")") {
		return map[string]interface{}{
			"regexp": strings.TrimSuffix(strings.SplitN(str, "(", 2)[1], ")"),
		}
	}

	return map[string]interface{}{
		"exact": str,
	}
}

func stringMatchToString(v interface{}) string {
	m := convert.ToMapInterface(v)

	exact, _ := m["exact"].(string)
	prefix, _ := m["prefix"].(string)
	regexp, _ := m["regexp"].(string)

	var result string
	if exact != "" {
		result = exact
	} else if prefix != "" {
		result = prefix + "*"
	} else if regexp != "" {
		result = "regex(" + regexp + ")"
	}

	return result
}
