package stack

import (
	"github.com/rancher/mapper/convert"
	"github.com/rancher/rio/cli/cmd/edit/edit"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type Editor struct {
	updater edit.Updater
}

func NewEditor(updater edit.Updater) Editor {
	return Editor{
		updater: updater,
	}
}

func (s Editor) Edit(obj runtime.Object) (bool, error) {
	stack := obj.(*riov1.Stack)

	return edit.Loop(nil, []byte(stack.Spec.Template), func(content []byte) error {
		stack.Spec.Template = string(content)
		m, err := convert.EncodeToMap(stack)
		if err != nil {
			return err
		}

		u := &unstructured.Unstructured{
			Object: m,
		}

		return s.updater.Update(u)
	})
}
