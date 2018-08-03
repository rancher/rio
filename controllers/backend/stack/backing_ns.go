package stack

import (
	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *stackController) createBackingNamespace(stack *v1beta1.Stack) (*v1beta1.Stack, error) {
	currentNs, err := s.namespaceLister.Get("", stack.Namespace)
	if err != nil {
		return nil, err
	}

	ns, err := s.namespaceLister.Get("", namespace.StackToNamespace(stack))
	if errors.IsNotFound(err) {
		ns = &k8sv1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:        namespace.StackToNamespace(stack),
				Annotations: map[string]string{},
			},
		}

		if project, ok := currentNs.Annotations[projectID]; ok {
			ns.Annotations[projectID] = project
		}

		_, err = s.namespaces.Create(ns)
	}

	v1beta1.StackConditionNamespaceCreated.True(stack)
	return stack, err
}
