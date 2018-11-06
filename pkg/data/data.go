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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func AddData(rContext *types.Context, inCluster bool) error {
	if err := addNameSpace(rContext); err != nil {
		return err
	}

	if err := apply.Apply(systemStacks(inCluster), nil, settings.RioSystemNamespace, "rio-system-stacks"); err != nil {
		return err
	}
	return nil
}

func systemStacks(inCluster bool) []runtime.Object {
	var result []runtime.Object

	result = append(result, stack("istio-crd", v1beta1.StackSpec{
		DisableMesh:               true,
		EnableKubernetesResources: true,
	}))

	result = append(result, stack("cert-manager-crd", v1beta1.StackSpec{
		DisableMesh:               true,
		EnableKubernetesResources: true,
	}))

	result = append(result, stack("storageclasses", v1beta1.StackSpec{
		DisableMesh:               true,
		EnableKubernetesResources: true,
	}))

	result = append(result, stack("local-storage", v1beta1.StackSpec{
		DisableMesh: true,
	}))

	result = append(result, stack("cert-manager", v1beta1.StackSpec{
		DisableMesh: true,
	}))

	if !inCluster {
		result = append(result, stack("coredns", v1beta1.StackSpec{
			DisableMesh: true,
		}))
	}

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

	return nil
}
