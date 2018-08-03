package data

import (
	"fmt"

	"github.com/rancher/rio/pkg/apply"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/pkg/space"
	"github.com/rancher/rio/stacks"
	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/staging/src/k8s.io/apimachinery/pkg/api/errors"
)

func AddData(rContext *types.Context, inCluster bool) error {
	if err := addNameSpace(rContext); err != nil {
		return err
	}

	return apply.Apply(systemStacks(inCluster), "rio-system-stacks", 0)
}

func systemStacks(inCluster bool) []runtime.Object {
	var result []runtime.Object

	if !inCluster {
		result = append(result, stack("coredns", v1beta1.StackSpec{
			DisableMesh: true,
		}))
	}

	result = append(result, stack("istio", v1beta1.StackSpec{
		DisableMesh:               true,
		EnableKubernetesResources: true,
	}))

	return result
}

func stack(name string, spec v1beta1.StackSpec) runtime.Object {
	s := &v1beta1.Stack{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rio.cattle.io/v1beta1",
			Kind:       "Stack",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: settings.RioSystemNamespace,
		},
		Spec: spec,
	}

	s.Spec = spec
	s.Spec.Template = stackData(name)

	return s
}

func stackData(name string) string {
	bytes, err := stacks.Asset(fmt.Sprintf("stacks/%s-stack.yml", name))
	if err != nil {
		panic("failed to load stack data for: " + name + " " + err.Error())
	}
	return string(bytes)
}

func addNameSpace(rContext *types.Context) error {
	_, err := rContext.Core.Namespaces("").Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: settings.RioSystemNamespace,
			Labels: map[string]string{
				space.SpaceLabel: "true",
			},
		},
	})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	_, err = rContext.Core.Namespaces("").Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: settings.RioDefaultNamespace,
			Labels: map[string]string{
				space.SpaceLabel: "true",
			},
		},
	})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	return nil
}
