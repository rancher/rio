package data

import (
	"fmt"

	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/stacks"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func systemStacks(inCluster bool) []runtime.Object {
	var result []runtime.Object

	if !inCluster {
		result = append(result, Stack("coredns", riov1.StackSpec{
			DisableMesh: true,
		}))
	}

	return result
}

func Stack(name string, spec riov1.StackSpec) runtime.Object {
	s := riov1.NewStack(settings.RioSystemNamespace, name, riov1.Stack{
		Spec: spec,
	})
	s.Spec.Template = StackData(name)

	return s
}

func StackData(name string) string {
	bytes, err := stacks.Asset(fmt.Sprintf("stacks/%s-stack.yaml", name))
	if err != nil {
		panic("failed to load stack data for: " + name + " " + err.Error())
	}
	return string(bytes)
}
