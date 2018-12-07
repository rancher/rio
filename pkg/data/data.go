package data

import (
	"fmt"

	"github.com/rancher/rio/pkg/apply"
	"github.com/rancher/rio/pkg/project"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/stacks"
	"github.com/rancher/rio/types"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func AddData(rContext *types.Context, inCluster bool) error {
	if err := addNameSpace(rContext); err != nil {
		return err
	}
	if err := addFeatures(rContext); err != nil {
		return err
	}

	if err := apply.Apply(systemStacks(inCluster), nil, settings.RioSystemNamespace, "rio-system-stacks"); err != nil {
		return err
	}

	return localStacks()
}

func systemStacks(inCluster bool) []runtime.Object {
	var result []runtime.Object

	result = append(result, Stack("istio-crd", riov1.StackSpec{
		DisableMesh:               true,
		EnableKubernetesResources: true,
	}))

	result = append(result, Stack("cert-manager-crd", riov1.StackSpec{
		DisableMesh:               true,
		EnableKubernetesResources: true,
	}))

	result = append(result, Stack("storageclasses", riov1.StackSpec{
		DisableMesh:               true,
		EnableKubernetesResources: true,
	}))

	result = append(result, Stack("local-storage", riov1.StackSpec{
		DisableMesh: true,
	}))

	result = append(result, Stack("cert-manager", riov1.StackSpec{
		DisableMesh: true,
		Answers: map[string]string{
			"CERT_MANAGER_IMAGE": settings.CertManagerImage.Get(),
		},
	}))

	if !inCluster {
		result = append(result, Stack("coredns", riov1.StackSpec{
			DisableMesh: true,
		}))
	}

	return result
}

func Stack(name string, spec riov1.StackSpec) runtime.Object {
	s := &riov1.Stack{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rio.cattle.io/v1",
			Kind:       "Stack",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: settings.RioSystemNamespace,
		},
		Spec: spec,
	}

	s.Spec = spec
	s.Spec.Template = StackData(name)

	return s
}

func StackData(name string) string {
	bytes, err := stacks.Asset(fmt.Sprintf("stacks/%s-stack.yml", name))
	if err != nil {
		panic("failed to load stack data for: " + name + " " + err.Error())
	}
	return string(bytes)
}

func addNameSpace(rContext *types.Context) error {
	_, err := rContext.Core.Namespace.Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: settings.RioSystemNamespace,
			Labels: map[string]string{
				project.ProjectLabel: "true",
			},
		},
	})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	return nil
}
