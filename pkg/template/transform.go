package template

import (
	"encoding/base64"
	"fmt"

	"github.com/rancher/norman/types/convert"
	"github.com/rancher/norman/types/convert/schemaconvert"
	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/pkg/pretty"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1/schema"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

func FromClientStack(stack *client.Stack) (*Template, error) {
	stackSchema := schema.Schemas.Schema(&schema.Version, client.StackType)
	internalStack := &v1beta1.Stack{}
	err := schemaconvert.ToInternal(stack, stackSchema, internalStack)
	if err != nil {
		return nil, err
	}

	return FromStack(internalStack)
}

func FromStack(stack *v1beta1.Stack) (*Template, error) {
	result := &Template{
		Namespace:       namespace.StackToNamespace(stack),
		Content:         []byte(stack.Spec.Template),
		Answers:         map[string]string{},
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
		result.Answers = stack.Spec.Answers
	}

	return result, nil
}

func (t *Template) ToInternalStack() (*v1beta1.InternalStack, error) {
	data, err := t.parseYAML()
	if err != nil {
		return nil, err
	}

	return pretty.ToInternalStack(data)
}

func (t *Template) ToClientStack() (*client.Stack, error) {
	result := &client.Stack{
		Answers:         t.Answers,
		Template:        string(t.Content),
		AdditionalFiles: map[string]string{},
	}

	for name, content := range t.AdditionalFiles {
		encoded := base64.StdEncoding.EncodeToString(content)
		result.AdditionalFiles[name] = encoded
	}

	for _, q := range t.Questions {
		newQ := client.Question{}
		if err := convert.ToObj(q, &newQ); err != nil {
			return nil, err
		}
		result.Questions = append(result.Questions, newQ)
	}

	return result, nil
}
