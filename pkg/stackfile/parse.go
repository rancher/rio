package stackfile

import (
	"encoding/base64"
	"fmt"
	"unicode/utf8"

	"github.com/rancher/mapper/convert"
	"github.com/rancher/mapper/values"
	"github.com/rancher/rio/pkg/template"
	"github.com/rancher/wrangler/pkg/yaml"
)

var builtinVars = []string{
	"NAMESPACE",
}

func ReadDir(dir string, answersFromEnv bool) (map[string]*StackFile, error) {
	templates, err := template.ReadDir("stack.yaml", dir, answersFromEnv)
	if err != nil {
		return nil, err
	}

	result := map[string]*StackFile{}
	for k, v := range templates {
		if v == nil {
			continue
		}
		result[k] = &StackFile{
			name:     k,
			Template: *v,
		}
	}

	return result, nil
}

func (t *StackFile) RequiredEnv() ([]string, error) {
	t.Template.BuiltinVars = builtinVars
	return t.Template.RequiredEnv()
}

func (t *StackFile) PopulateAnswersFromEnv() error {
	return t.PopulateAnswersFromEnv()
}

func (t *StackFile) RequiredFiles() ([]string, error) {
	content, err := t.Template.Parse(nil)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{}
	if err := yaml.Unmarshal(content, &data); err != nil {
		return nil, err
	}

	var result []string
	for _, v := range fileReferences("configs", data) {
		result = append(result, v)
	}

	return result, nil
}

func fileReferences(key string, data map[string]interface{}) map[string]string {
	result := map[string]string{}

	configs := convert.ToMapInterface(data[key])
	for k, v := range configs {
		str, _ := convert.ToMapInterface(v)["file"].(string)
		if str != "" {
			result[k] = str
		}
	}

	return result
}

func (t *StackFile) replaceFileReferences(data map[string]interface{}, key string) error {
	for configKey, file := range fileReferences(key, data) {
		bytes, ok := t.AdditionalFiles[convert.ToString(file)]
		if !ok {
			return fmt.Errorf("missing file: %v", file)
		}

		if utf8.Valid(bytes) {
			values.PutValue(data, string(bytes), key, configKey, "content")
		} else {
			values.PutValue(data, base64.StdEncoding.EncodeToString(bytes), key, configKey, "encoded")
		}
	}

	return nil
}

func (t *StackFile) Validate() error {
	return t.Template.Validate()
}
