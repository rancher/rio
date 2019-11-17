package template

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/drone/envsubst"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/template/gotemplate"
	"github.com/rancher/wrangler/pkg/merr"
	"github.com/rancher/wrangler/pkg/yaml"
)

type templateFile struct {
	Meta v1.TemplateMeta `json:"template"`
}

func (t *Template) Questions() ([]v1.Question, error) {
	tf, err := t.readTemplate()
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

func (t *Template) afterTemplate(content []byte) []byte {
	found := false

	result := bytes.Buffer{}
	scan := bufio.NewScanner(bytes.NewReader(content))
	for scan.Scan() {
		if strings.HasPrefix(string(scan.Bytes()), "template:") {
			found = true
		}
		if found {
			result.Write(scan.Bytes())
			result.WriteRune('\n')
		}
	}

	if found {
		return result.Bytes()
	}
	return content
}

func (t *Template) readTemplate() (*templateFile, error) {
	content, err := gotemplate.Apply(t.afterTemplate(t.Content), nil)
	if err != nil {
		return nil, nil
	}

	return t.readTemplateFile(content)
}

func (t *Template) parseContent(answersCB AnswerCallback) ([]byte, error) {
	template, err := t.readTemplate()
	if err != nil {
		return nil, err
	}

	var (
		callbackErrs []error
		answers      = map[string]string{}
		evaled       = string(t.Content)
	)

	if template.Meta.EnvSubst {
		evaled, err = envsubst.Eval(evaled, func(key string) string {
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
	}

	for _, q := range template.Meta.Questions {
		if answersCB == nil {
			answers[q.Variable] = q.Default
			break
		}
		val, err := answersCB(q.Variable, template.Meta.Questions)
		if err != nil {
			callbackErrs = append(callbackErrs, err)
		}
		answers[q.Variable] = val
	}
	if err != nil {
		return nil, err
	} else if len(callbackErrs) > 0 {
		return nil, merr.NewErrors(callbackErrs...)
	}

	if template.Meta.GoTemplate {
		return gotemplate.Apply([]byte(evaled), answers)
	}

	return []byte(evaled), nil
}
