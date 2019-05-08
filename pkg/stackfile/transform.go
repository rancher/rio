package stackfile

import (
	"encoding/base64"
	"fmt"

	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/template"
)

func FromStack(stack *v1.Stack) (*StackFile, error) {
	result := &StackFile{
		name: stack.Name,
		Template: template.Template{
			Content: []byte(stack.Spec.Template),
			Answers: map[string]string{},
		},
		AdditionalFiles: map[string][]byte{},
	}

	for name, value := range stack.Spec.AdditionalFiles {
		content, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template [%s]: %v", name, err)
		}
		result.AdditionalFiles[name] = content
	}

	if stack.Spec.Answers != nil {
		result.Template.Answers = stack.Spec.Answers
	}

	return result, nil
}

func (t *StackFile) ToStackResources() (*v1.StackFile, error) {
	return nil, nil
}

func (t *StackFile) ToStackSpec() v1.StackSpec {
	s := v1.StackSpec{
		Template:  string(t.Template.Content),
		Answers:   t.Template.Answers,
		Questions: t.Template.Questions,
	}

	for name, value := range t.AdditionalFiles {
		s.AdditionalFiles[name] = base64.StdEncoding.EncodeToString(value)
	}

	return s
}
