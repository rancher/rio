package stackdef

import (
	"context"

	"github.com/rancher/rio/pkg/deploy/stackdef"
	"github.com/rancher/rio/types"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"github.com/rancher/types/apis/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) {
	s := &stackController{
		namespaceLister: rContext.Core.Namespace.Cache(),
	}
	rContext.Rio.Stack.OnRemove(ctx, "stackdef-controller", s.Remove)
	rContext.Rio.Stack.OnChange(ctx, "stackdef-controller", s.Updated)
}

type stackController struct {
	namespaceLister v1.NamespaceClientCache
}

func (s *stackController) Remove(obj *riov1.Stack) (runtime.Object, error) {
	return obj, stackdef.Remove(obj)
}

func (s *stackController) Updated(stack *riov1.Stack) (runtime.Object, error) {
	ns, err := s.namespaceLister.Get("", stack.Namespace)
	if err != nil {
		return nil, err
	}
	_, err = riov1.StackConditionDefined.Do(stack, func() (runtime.Object, error) {
		return stack, stackdef.Deploy(ns, stack)
	})
	riov1.PendingCondition.True(stack)
	return stack, err
}
