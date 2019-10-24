package stack

import (
	"io/ioutil"

	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	adminv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/riofile"
	"github.com/rancher/rio/pkg/template"
	"github.com/rancher/rio/stacks"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/objectset"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

type SystemStack struct {
	k8s    dynamic.Interface
	apply  apply.Apply
	stacks adminv1controller.SystemStackClient
	name   string
	Stack
}

func NewSystemStack(apply apply.Apply, stacks adminv1controller.SystemStackClient, systemNamespace string, name string) *SystemStack {
	setID := "system-stack-" + name
	s := &SystemStack{
		apply:  apply.WithSetID(setID).WithDefaultNamespace(systemNamespace).WithDynamicLookup(),
		stacks: stacks,
		name:   name,
		Stack:  Stack{},
	}
	if stacks != nil {
		stack, err := stacks.Get(name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			stack, _ = stacks.Create(&adminv1.SystemStack{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
			})
		}
		s.apply = s.apply.WithSetOwnerReference(true, true).WithOwner(stack)
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
	if s.stacks != nil {
		return s.stacks.Delete(s.name, &metav1.DeleteOptions{})
	}
	return s.apply.Apply(nil)
}

func (s *SystemStack) WithApply(apply apply.Apply) {
	s.apply = apply
}

func (s *SystemStack) content() ([]byte, error) {
	if constants.DevMode {
		return ioutil.ReadFile("stacks/" + s.name + "-stack.yaml")
	}
	return stacks.Asset("stacks/" + s.name + "-stack.yaml")
}
