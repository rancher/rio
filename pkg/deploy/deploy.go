package deploy

import (
	"fmt"

	"github.com/rancher/rio/pkg/apply"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Deploy(namespace string, stack *StackResources, injectors ...apply.ConfigInjector) error {
	configs, objects, err := configs(nil, stack, namespace)
	if err != nil {
		return err
	}

	objects, err = volumes(objects, stack, namespace)
	if err != nil {
		return err
	}

	objects, err = services(objects, configs, stack, namespace)
	if err != nil {
		return err
	}

	namespaced, global, err := splitObjects(objects)
	if err != nil {
		return err
	}

	if len(global) > 0 {
		if err := apply.Apply(global, "stackdeploy-global-"+namespace, 0, injectors...); err != nil {
			return err
		}
	}

	return apply.Apply(namespaced, "stackdeploy-"+namespace, 0, injectors...)
}

func splitObjects(objects []runtime.Object) ([]runtime.Object, []runtime.Object, error) {
	var (
		ns     []runtime.Object
		global []runtime.Object
	)

	for _, obj := range objects {
		metaObj, ok := obj.(v1.Object)
		if !ok {
			return nil, nil, fmt.Errorf("resource type is not a meta object")
		}

		if len(metaObj.GetNamespace()) > 0 {
			ns = append(ns, obj)
		} else {
			global = append(global, obj)
		}
	}

	return ns, global, nil
}
