package stack

import (
	"context"

	"github.com/rancher/rio/modules/stack/controllers/stack/populate"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/generated/controllers/core/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	sc := stackobject.NewGeneratingController(ctx, rContext, "stack-stack", rContext.Rio.Rio().V1().Stack())
	sc.Apply = rContext.Apply.WithSetID("stack-stack").
		WithCacheTypes(rContext.Rio.Rio().V1().Service(),
			rContext.Rio.Rio().V1().Config(),
			rContext.Rio.Rio().V1().Volume(),
			rContext.Rio.Rio().V1().Router(),
			rContext.Rio.Rio().V1().ExternalService(),
			rContext.Storage.Storage().V1().StorageClass(),
			rContext.Ext.Apiextensions().V1beta1().CustomResourceDefinition())

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
