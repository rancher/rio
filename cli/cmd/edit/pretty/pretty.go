package pretty

import (
	"github.com/rancher/rio/cli/cmd/edit/edit"
	"github.com/rancher/rio/pkg/riofile"
	"k8s.io/apimachinery/pkg/runtime"
)

type Editor struct {
	updater edit.Updater
}

func NewEditor(updator edit.Updater) Editor {
	return Editor{
		updater: updator,
	}
}

func (r Editor) Edit(obj runtime.Object) (bool, error) {
	content, err := riofile.RenderObject(obj)
	if err != nil {
		return false, err
	}

	return edit.Loop(nil, content, func(modifiedContent []byte) error {
		modifiedObj, err := riofile.Update(obj, modifiedContent)
		if err != nil {
			return err
		}

		return r.updater.Update(modifiedObj)
	})
}
