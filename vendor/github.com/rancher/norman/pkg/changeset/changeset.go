package changeset

import (
	"strings"

	"github.com/rancher/norman/controller"
	"k8s.io/apimachinery/pkg/runtime"
)

type Key struct {
	Namespace string
	Name      string
}

type ControllerProvider interface {
	Generic() controller.GenericController
}

type Enqueuer func(namespace, name string)

type Resolver func(namespace, name string, obj runtime.Object) ([]Key, error)

func Watch(name string, resolve Resolver, enq Enqueuer, controllers ...ControllerProvider) {
	for _, c := range controllers {
		watch(name, enq, resolve, c.Generic())
	}
}

func watch(name string, enq Enqueuer, resolve Resolver, genericController controller.GenericController) {
	genericController.AddHandler(name, func(key string) error {
		obj, exists, err := genericController.Informer().GetStore().GetByKey(key)
		if err != nil {
			return err
		}

		if !exists {
			obj = nil
		}

		var (
			ns   string
			name string
		)

		parts := strings.SplitN(key, "/", 2)
		if len(parts) == 2 {
			ns = parts[0]
			name = parts[1]
		} else {
			name = parts[0]
		}

		ro, ok := obj.(runtime.Object)
		if !ok {
			ro = nil
		}

		keys, err := resolve(ns, name, ro)
		if err != nil {
			return err
		}

		for _, key := range keys {
			if key.Name != "" {
				enq(key.Namespace, key.Name)
			}
		}

		return nil
	})
}
