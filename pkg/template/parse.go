package template

import (
	"github.com/drone/envsubst"
	"github.com/rancher/mapper"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/template/gotemplate"
	"github.com/rancher/wrangler/pkg/yaml"
)

type templateFile struct {
	Meta v1.TemplateMeta `json:"template"`
}

func (t *Template) Questions() ([]v1.Question, error) {
	content, err := t.parseContent(nil)
	if err != nil {
		return nil, err
	}

	tf, err := t.readTemplateFile(content)
	if err != nil {
		return nil, err
	}

	return tf.Meta.Questions, nil
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

	template, err := t.readTemplateFile(t.Content)
	if err != nil {
		return nil, err
	}

	for _, q := range template.Meta.Questions {
		names[q.Variable] = true
	}

	var result []string
	for key := range names {
		result = append(result, key)
	}

	return result, nil
}

func (t *Template) readTemplateFile(content []byte) (*templateFile, error) {
	templateFile := &templateFile{}
	return templateFile, yaml.Unmarshal(content, templateFile)
}

func (t *Template) Parse(answers AnswerCallback) ([]byte, error) {
	return t.parseContent(answers)
}

func (t *Template) Validate() error {
	content, err := t.parseContent(nil)
	if err != nil {
		return err
	}

	_, err = t.readTemplateFile(content)
	return err
}

func (t *Template) parseContent(answersCB AnswerCallback) ([]byte, error) {
	content, err := gotemplate.Apply(t.Content, nil)
	if err != nil {
		return nil, err
	}

	template, err := t.readTemplateFile(content)
	if err != nil {
		return nil, err
	}

	var (
		callbackErrs []error
		answers      = map[string]string{}
	)

	evaled, err := envsubst.Eval(string(t.Content), func(key string) string {
		if answersCB == nil {
			return ""
		}
		val, err := answersCB(key, template.Meta.Questions)
		if err != nil {
			callbackErrs = append(callbackErrs, err)
		}
		answers[key] = val
		return val
	})
	if err != nil {
		return nil, err
	} else if len(callbackErrs) > 0 {
		return nil, mapper.NewErrors(callbackErrs...)
	}

	return gotemplate.Apply([]byte(evaled), answers)
}
