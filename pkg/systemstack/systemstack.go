package systemstack

import (
	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/pkg/template"
	"github.com/rancher/rio/stacks"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"github.com/rancher/types/apis/management.cattle.io/v3"
)

type SystemStack struct {
	processor objectset.Processor
	spec      v1.StackSpec
	name      string
}

func NewSystemStack(stacksClient v1.StackClient, name string, spec v1.StackSpec) *SystemStack {
	return &SystemStack{
		processor: objectset.NewProcessor("system-stack-" + name).
			Client(stacksClient),
		spec: spec,
		name: name,
	}
}

func (s *SystemStack) Questions() ([]v3.Question, error) {
	content, err := s.content()
	if err != nil {
		return nil, err
	}

	t := template.Template{
		Content: content,
	}
	if err := t.Validate(); err != nil {
		return nil, err
	}

	return t.Questions, nil
}

func (s *SystemStack) content() ([]byte, error) {
	return stacks.Asset("stacks/" + s.name + "-stack.yaml")
}

func (s *SystemStack) Deploy(answers map[string]string) error {
	content, err := s.content()
	if err != nil {
		return err
	}

	stack := v1.NewStack(settings.RioSystemNamespace, s.name, v1.Stack{
		Spec: s.spec,
	})

	for k, v := range answers {
		if stack.Spec.Answers == nil {
			stack.Spec.Answers = map[string]string{}
		}
		stack.Spec.Answers[k] = v
	}
	stack.Spec.Template = string(content)

	return s.processor.NewDesiredSet(nil, objectset.NewObjectSet().Add(stack)).Apply()
}

func (s *SystemStack) Remove() error {
	return s.processor.Remove(nil)
}
