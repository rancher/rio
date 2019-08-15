package raw

import (
	"encoding/json"

	"github.com/rancher/rio/cli/cmd/edit/edit"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

type Editor struct {
	updater edit.Updater
}

func NewRawEditor(updator edit.Updater) Editor {
	return Editor{
		updater: updator,
	}
}

func (r Editor) Edit(obj runtime.Object) (bool, error) {
	m, err := json.Marshal(obj)
	if err != nil {
		return false, err
	}
	str, err := yaml.JSONToYAML(m)
	if err != nil {
		return false, err
	}

	return edit.Loop(nil, str, func(content []byte) error {
		m := make(map[string]interface{})
		if err := yaml.Unmarshal(content, &m); err != nil {
			return err
		}
		obj := &unstructured.Unstructured{
			Object: m,
		}
		obj.SetGroupVersionKind(r.updater.GetGvk())

		return r.updater.Update(obj)
	})
}
