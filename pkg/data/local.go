package data

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/rancher/rio/pkg/apply"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/pkg/space"
	"github.com/rancher/rio/pkg/template"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var projectsDir = "./projects"

func localStacks() error {
	data := &apply.Data{
		GroupID: "local-stacks",
	}

	if err := readDir(data, projectsDir); err != nil {
		return err
	}
	if err := readDir(data, settings.LocalStacksDir.Get()); err != nil {
		return err
	}

	return data.Apply()
}

func readDir(data *apply.Data, base string) error {
	entries, err := ioutil.ReadDir(base)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	for _, dir := range entries {
		if !dir.IsDir() {
			continue
		}

		templates, err := template.ReadDir(filepath.Join(base, dir.Name()), true)
		if err != nil {
			return err
		}

		if len(templates) == 0 {
			continue
		}

		data.Add("", "v1", "Namespace", map[string]runtime.Object{
			dir.Name(): &v1.Namespace{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Namespace",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: dir.Name(),
					Labels: map[string]string{
						space.SpaceLabel:              "true",
						"field.cattle.io/displayName": dir.Name(),
					},
				},
			},
		})

		for name, template := range templates {
			stack := template.ToStack(dir.Name(), name)
			data.Add(stack.Namespace, "rio.cattle.io/v1beta1", "Stack", map[string]runtime.Object{
				stack.Name: stack,
			})
		}
	}

	return nil
}
