package yaml

import (
	"github.com/rancher/norman/types/convert"
	"gopkg.in/yaml.v2"
)

func Parse(content []byte) (map[string]interface{}, error) {
	data := map[interface{}]interface{}{}
	err := yaml.Unmarshal(content, &data)
	if err != nil {
		return nil, err
	}

	return convertMap(data, false), nil
}

func ConvertYAMLKeys(data map[string]interface{}) map[string]interface{} {
	return convertMapString(data, true)
}

func convertSlice(data []interface{}, toYAML bool) []interface{} {
	var result []interface{}
	for _, obj := range data {
		result = append(result, convertValue(obj, toYAML))
	}
	return result
}

func convertMapString(data map[string]interface{}, toYAML bool) map[string]interface{} {
	result := map[string]interface{}{}
	for k, v := range data {
		result[convertKey(k, toYAML)] = convertValue(v, toYAML)
	}
	return result
}

func convertMap(data map[interface{}]interface{}, toYAML bool) map[string]interface{} {
	result := map[string]interface{}{}
	for k, v := range data {
		result[convertKey(k, toYAML)] = convertValue(v, toYAML)
	}
	return result
}

func convertValue(val interface{}, toYAML bool) interface{} {
	switch v := val.(type) {
	case map[string]interface{}:
		return convertMapString(v, toYAML)
	case map[interface{}]interface{}:
		return convertMap(v, toYAML)
	case []interface{}:
		return convertSlice(v, toYAML)
	default:
		return val
	}
}

func convertKey(k interface{}, toYAML bool) string {
	str := convert.ToString(k)
	if toYAML {
		str = convert.ToYAMLKey(str)
	}
	return str
}
