package stacknamespace

import (
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	corev1controller "github.com/rancher/rio/pkg/generated/controllers/core/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	stackNamespaceLabel = "stack.rio.cattle.io/namespace"
	stackNameLabel      = "stack.rio.cattle.io/name"
)

func SetStackLabels(stack *v1.Stack, ns *corev1.Namespace) {
	if ns.Labels == nil {
		ns.Labels = map[string]string{}
	}
	ns.Labels[stackNameLabel] = stack.Name
	ns.Labels[stackNamespaceLabel] = stack.Namespace
}

func GetStackFromNamespace(ns *corev1.Namespace) (namespace, name string) {
	return ns.Labels[stackNamespaceLabel], ns.Labels[stackNameLabel]
}

func GetStack(obj metav1.Object, stackCache riov1controller.StackCache, namespaces corev1controller.NamespaceCache) (*v1.Stack, error) {
	if s, ok := obj.(*v1.Stack); ok {
		return s, nil
	}
	nsObj, err := namespaces.Get(obj.GetNamespace())
	if err != nil {
		return nil, err
	}

	ns, name := GetStackFromNamespace(nsObj)
	return stackCache.Get(ns, name)
}
