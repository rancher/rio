package stackns

import (
	"context"

	"github.com/rancher/rio/modules/stack/controllers/stackns/populate"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/generated/controllers/core/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	sc := stackobject.NewGeneratingController(ctx, rContext, "stack-stackns", rContext.Rio.Rio().V1().Stack())
	sc.Apply = sc.Apply.WithCacheTypes(rContext.Core.Core().V1().Namespace())

	s := &stackController{
		namespaceLister: rContext.Core.Core().V1().Namespace().Cache(),
	}
	sc.Populator = s.populate

	return nil
}

type stackController struct {
	namespaceLister v1.NamespaceCache
}

func (s *stackController) populate(obj runtime.Object, stack *riov1.Stack, os *objectset.ObjectSet) error {
	ns, err := s.namespaceLister.Get(stack.Namespace)
	if err != nil {
		return err
	}

	riov1.PendingCondition.True(stack)
	return populate.Stack(ns, stack, os)
}
