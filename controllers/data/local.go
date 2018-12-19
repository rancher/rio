package data

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/pkg/project"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/pkg/template"
	v12 "github.com/rancher/types/apis/core/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var projectsDir = "./projects"

func localStacks(os *objectset.ObjectSet) error {
	if err := readDir(os, projectsDir); err != nil {
		return err
	}

	return readDir(os, settings.LocalStacksDir.Get())
}

func readDir(objSet *objectset.ObjectSet, base string) error {
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

		ns := v12.NewNamespace("", dir.Name(), v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					project.ProjectLabel:          "true",
					"field.cattle.io/displayName": dir.Name(),
				},
			},
		})
		objSet.Add(ns)

		for name, template := range templates {
			stack := template.ToStack(dir.Name(), name)
			objSet.Add(stack)
		}
	}

	return nil
}
