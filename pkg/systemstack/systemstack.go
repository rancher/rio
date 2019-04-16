package systemstack

import (
	"bytes"

	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/riofile"
	"github.com/rancher/rio/pkg/template"
	"github.com/rancher/rio/stacks"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/objectset"
)

type SystemStack struct {
	namespace string
	apply     apply.Apply
	name      string
}

func NewSystemStack(apply apply.Apply, systemNamespace string, name string) *SystemStack {
	return &SystemStack{
		namespace: systemNamespace,
		apply:     apply.WithSetID("system-stack-" + name),
		name:      name,
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

	return t.Questions()
}

func (s *SystemStack) content() ([]byte, error) {
	return stacks.Asset("stacks/" + s.name + "-stack.yaml")
}

func (s *SystemStack) Deploy(answers map[string]string) error {
	content, err := s.content()
	if err != nil {
		return err
	}

	rf, err := riofile.Parse(bytes.NewBuffer(content), template.AnswersFromMap(answers))
	if err != nil {
		return err
	}

	os := objectset.NewObjectSet()
	os.Add(rf.Objects()...)
	return s.apply.Apply(os)
}

func (s *SystemStack) Remove() error {
	return s.apply.Apply(nil)
}
