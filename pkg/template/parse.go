package template

import (
	"bytes"
	"os"

	"github.com/drone/envsubst"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/template/gotemplate"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type templateFile struct {
	Meta      v1.TemplateMeta `json:"meta"`
	Questions []v1.Question   `json:"questions"`
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

	for _, b := range t.BuiltinVars {
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

func (t *Template) PopulateAnswersFromEnv() error {
	keys, err := t.RequiredEnv()
	if err != nil {
		return err
	}

	for _, key := range keys {
		value := os.Getenv(key)
		if value != "" {
			t.Answers[key] = value
		}
	}

	return nil
}

func (t *Template) readTemplateFile(content []byte) (*templateFile, error) {
	templateFile := &templateFile{}
	yamlParser := yaml.NewYAMLToJSONDecoder(bytes.NewBuffer(content))
	return templateFile, yamlParser.Decode(templateFile)
}

func (t *Template) Parse(answers map[string]string) ([]byte, error) {
	return t.parseContent(answers)

}

func (t *Template) Validate() error {
	content, err := t.parseContent(nil)
	if err != nil {
		return err
	}

	templateFile, err := t.readTemplateFile(content)
	if err != nil {
		return err
	}

	t.Meta = templateFile.Meta
	t.Questions = templateFile.Questions
	return nil
}

func (t *Template) parseYAML() ([]byte, error) {
	answers := map[string]string{}
	for _, q := range t.Questions {
		answers[q.Variable] = q.Default
	}
	for k, v := range t.Answers {
		answers[k] = v
	}
	return t.parseContent(answers)
}

func (t *Template) parseContent(answers map[string]string) ([]byte, error) {
	evaled, err := envsubst.Eval(string(t.Content), func(key string) string {
		return answers[key]
	})
	if err != nil {
		return nil, err
	}

	return gotemplate.Apply([]byte(evaled), t.Answers)
}
