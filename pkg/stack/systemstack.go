package stack

import (
	"io/ioutil"

	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/riofile"
	"github.com/rancher/rio/pkg/template"
	"github.com/rancher/rio/stacks"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/objectset"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
)

type SystemStack struct {
	k8s   dynamic.Interface
	apply apply.Apply
	name  string
	Stack
}

func NewSystemStack(apply apply.Apply, systemNamespace string, name string) *SystemStack {
	setID := "system-stack-" + name
	s := &SystemStack{
		apply: apply.WithSetID(setID).WithDefaultNamespace(systemNamespace),
		name:  name,
		Stack: Stack{},
	}
	contents, err := s.content()
	if err != nil {
		logrus.Fatal(err)
	}
	s.contents = contents
	return s
}

func (s *SystemStack) Deploy(answers map[string]string) error {
	content, err := s.content()
	if err != nil {
		return err
	}

	rf, err := riofile.Parse(content, template.AnswersFromMap(answers))
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

func (s *SystemStack) content() ([]byte, error) {
	if constants.DevMode != "" {
		return ioutil.ReadFile("stacks/" + s.name + "-stack.yaml")
	}
	return stacks.Asset("stacks/" + s.name + "-stack.yaml")
}
