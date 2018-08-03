package stack

import (
	"context"

	"github.com/rancher/rio/pkg/apply"
	"github.com/rancher/rio/pkg/namespace"
	template2 "github.com/rancher/rio/pkg/template"
	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"github.com/rancher/types/apis/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	projectID = "field.cattle.io/projectId"
)

func Register(ctx context.Context, rContext *types.Context) {
	s := &stackController{
		namespaceLister: rContext.Core.Namespaces("").Controller().Lister(),
		namespaces:      rContext.Core.Namespaces(""),
	}
	rContext.Rio.Stacks("").AddLifecycle("stack-controller", s)
}

type stackController struct {
	namespaceLister v1.NamespaceLister
	namespaces      v1.NamespaceInterface
}

func (s *stackController) Create(obj *v1beta1.Stack) (*v1beta1.Stack, error) {
	return obj, nil
}

func (s *stackController) Remove(obj *v1beta1.Stack) (*v1beta1.Stack, error) {
	err := s.namespaces.Delete(namespace.StackToNamespace(obj), nil)
	if errors.IsNotFound(err) {
		return obj, nil
	}
	return obj, err
}

func (s *stackController) Updated(stack *v1beta1.Stack) (*v1beta1.Stack, error) {
	_, err := v1beta1.StackConditionDefined.Do(stack, func() (runtime.Object, error) {
		return s.define(stack)
	})
	return stack, err
}

func (s *stackController) define(stack *v1beta1.Stack) (*v1beta1.Stack, error) {
	stack, err := s.createBackingNamespace(stack)
	if err != nil {
		return stack, err
	}

	internalStack, err := s.parseStack(stack)
	if err != nil {
		// if parsing fails we don't return err because it's a user error
		return stack, nil
	}

	ns := namespace.StackToNamespace(stack)

	if stack.Spec.EnableKubernetesResources {
		if err := deployK8sResources(stack.Name, ns, internalStack); err != nil {
			return stack, err
		}
	}

	objects := s.gatherObjects(ns, stack, internalStack)

	err = apply.Apply(objects, "stack-"+stack.Name, stack.Generation)
	return stack, err
}

func (s *stackController) parseStack(stack *v1beta1.Stack) (*v1beta1.InternalStack, error) {
	var internalStack *v1beta1.InternalStack

	_, err := v1beta1.StackConditionParsed.Do(stack, func() (runtime.Object, error) {
		t, err := template2.FromStack(stack)
		if err != nil {
			return nil, err
		}

		if err := t.Validate(); err != nil {
			return nil, err
		}

		internalStack, err = t.ToInternalStack()
		return nil, err
	})

	return internalStack, err
}
