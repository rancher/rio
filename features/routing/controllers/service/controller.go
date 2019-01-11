package service

import (
	"context"

	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/features/routing/controllers/service/populate"
	"github.com/rancher/rio/features/routing/controllers/util"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "routing-service", rContext.Rio.Service)
	c.Processor.Client(
		rContext.Networking.DestinationRule,
		rContext.Networking.VirtualService)

	sh := &serviceHandler{
		serviceClient: rContext.Rio.Service,
		serviceCache:  rContext.Rio.Service.Cache(),
	}

	// as a side effect of the stack service controller, all changes to revisions will enqueue the parent service

	c.Populator = sh.populate
	return nil
}

type serviceHandler struct {
	serviceClient riov1.ServiceClient
	serviceCache  riov1.ServiceClientCache
}

func (s *serviceHandler) populate(obj runtime.Object, stack *riov1.Stack, os *objectset.ObjectSet) error {
	if stack != nil && stack.Spec.DisableMesh {
		return nil
	}

	if err := util.WaitForClusterDomain(); err != nil {
		return err
	}

	service := obj.(*riov1.Service)
	services, err := s.serviceCache.List(service.Namespace, labels.Everything())
	if err != nil {
		return err
	}

	return populate.DestinationRulesAndVirtualServices(stack, services, service, os)
}
