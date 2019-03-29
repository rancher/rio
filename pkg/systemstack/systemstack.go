package systemstack

import (
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/template"
	"github.com/rancher/rio/stacks"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/objectset"
)

type SystemStack struct {
	namespace string
	apply     apply.Apply
	spec      v1.StackSpec
	name      string
}

func NewSystemStack(apply apply.Apply, systemNamespace string, stacksClient riov1controller.StackController, name string, spec v1.StackSpec) *SystemStack {
	return &SystemStack{
		namespace: systemNamespace,
		apply: apply.WithSetID("system-stack-" + name).
			WithCacheTypes(stacksClient),
		spec: spec,
		name: name,
	}
}

func (s *SystemStack) Questions() ([]v1.Question, error) {
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

	stack := v1.NewStack(s.namespace, s.name, v1.Stack{
		Spec: s.spec,
	})

	for k, v := range answers {
		if stack.Spec.Answers == nil {
			stack.Spec.Answers = map[string]string{}
		}
		stack.Spec.Answers[k] = v
	}
	stack.Spec.Template = string(content)

	return s.apply.Apply(objectset.NewObjectSet().Add(stack))
}

func (s *SystemStack) Remove() error {
	return s.apply.Apply(nil)
}
