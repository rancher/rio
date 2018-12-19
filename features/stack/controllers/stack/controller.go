package stack

import (
	"context"

	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/features/stack/controllers/stack/populate"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"github.com/rancher/types/apis/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	sc := stackobject.NewGeneratingController(ctx, rContext, "stack-stack", rContext.Rio.Stack)
	sc.Processor.Client(rContext.Rio.Service,
		rContext.Rio.Config,
		rContext.Rio.Volume,
		rContext.Rio.RouteSet,
		rContext.Rio.ExternalService,
		rContext.Core.Namespace,
		rContext.Storage.StorageClass,
		rContext.Ext.CustomResourceDefinition)

	s := &stackController{
		namespaceLister: rContext.Core.Namespace.Cache(),
	}
	sc.Populator = s.populate

	return nil
}

type stackController struct {
	namespaceLister v1.NamespaceClientCache
}

func (s *stackController) populate(obj runtime.Object, stack *riov1.Stack, os *objectset.ObjectSet) error {
	ns, err := s.namespaceLister.Get("", stack.Namespace)
	if err != nil {
		return err
	}

	riov1.PendingCondition.True(stack)
	return populate.Stack(ns, stack, os)
}
