package service

import (
	"context"

	"github.com/rancher/rio/features/routing/controllers/service/populate"
	"github.com/rancher/rio/features/routing/controllers/util"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "routing-service", rContext.Rio.Rio().V1().Service())
	c.Apply.WithStrictCaching().
		WithCacheTypes(rContext.Networking.Networking().V1alpha3().DestinationRule(),
			rContext.Networking.Networking().V1alpha3().VirtualService())

	sh := &serviceHandler{
		systemNamespace:      rContext.SystemNamespace,
		serviceClient:        rContext.Rio.Rio().V1().Service(),
		serviceCache:         rContext.Rio.Rio().V1().Service().Cache(),
		externalServiceCache: rContext.Rio.Rio().V1().ExternalService().Cache(),
	}

	// as a side effect of the stack service controller, all changes to revisions will enqueue the parent service

	c.Populator = sh.populate
	return nil
}

type serviceHandler struct {
	systemNamespace      string
	serviceClient        v1.ServiceClient
	serviceCache         v1.ServiceCache
	externalServiceCache v1.ExternalServiceCache
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
	return populate.DestinationRulesAndVirtualServices(s.systemNamespace, stack, services, service, os)
}
