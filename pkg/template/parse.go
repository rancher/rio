package template

import (
	"encoding/base64"
	"fmt"
	"unicode/utf8"

	"github.com/drone/envsubst"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/norman/types/values"
	"github.com/rancher/rio/pkg/pretty"
	"github.com/rancher/rio/pkg/template/gotemplate"
	"github.com/rancher/rio/pkg/yaml"
)

var builtinVars = []string{
	"NAMESPACE",
}

func (t *Template) RequiredEnv() ([]string, error) {
	names := map[string]bool{}
	_, err := envsubst.Eval(string(t.Content), func(in string) string {
		names[in] = true
		return in
	})
	if err != nil {
		return nil, err
	}

	for _, b := range builtinVars {
		delete(names, b)
	}

	for _, q := range t.Questions {
		names[q.Variable] = true
	}

	var result []string
	for key := range names {
		result = append(result, key)
	}

	return result, nil
}

func (t *Template) RequiredFiles() ([]string, error) {
	content, err := t.parseContent(nil)
	if err != nil {
		return nil, err
	}

	data, err := yaml.Parse(content)
	if err != nil {
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

func (t *Template) replaceFileReferences(data map[string]interface{}, key string) error {
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

func (t *Template) parseYAMLAndIncludeFiles(content []byte) (map[string]interface{}, error) {
	data, err := yaml.Parse(content)
	if err != nil {
		return nil, err
	}

	if err := t.replaceFileReferences(data, "configs"); err != nil {
		return nil, err
	}

	return data, nil
}

func (t *Template) Validate() error {
	content, err := t.parseContent(nil)
	if err != nil {
		return err
	}

	data, err := t.parseYAMLAndIncludeFiles(content)
	if err != nil {
		return err
	}

	stack, err := pretty.ToNormalizedStack(data)
	if err != nil {
		return err
	}

	t.Meta = stack.Meta
	t.Questions = stack.Questions
	return nil
}

func (t *Template) parseYAML() (map[string]interface{}, error) {
	answers := map[string]string{}
	for _, q := range t.Questions {
		answers[q.Variable] = q.Default
	}
	for k, v := range t.Answers {
		answers[k] = v
	}
	content, err := t.parseContent(answers)
	if err != nil {
		return nil, err
	}

	return t.parseYAMLAndIncludeFiles(content)
}

func (t *Template) parseContent(answers map[string]string) ([]byte, error) {
	evaled, err := envsubst.Eval(string(t.Content), func(key string) string {
		if key == "NAMESPACE" {
			return t.Namespace
		}
		return answers[key]
	})
	if err != nil {
		return nil, err
	}

	return gotemplate.Apply([]byte(evaled), t.Answers)
}
