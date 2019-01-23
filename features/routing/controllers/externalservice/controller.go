package externalservice

import (
	"context"

	"github.com/rancher/norman/pkg/kv"
	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/features/routing/controllers/externalservice/populate"
	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/pkg/serviceset"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "routing-external-service", rContext.Rio.ExternalService)
	c.Processor.Client(rContext.Networking.ServiceEntry, rContext.Networking.VirtualService)

	p := populator{
		serviceCache: rContext.Rio.Service.Cache(),
	}

	c.Populator = p.populate
	return nil
}

type populator struct {
	serviceCache riov1.ServiceClientCache
}

func (p populator) populate(obj runtime.Object, stack *riov1.Stack, os *objectset.ObjectSet) error {
	if err := populate.ServiceEntry(obj.(*riov1.ExternalService), stack, os); err != nil {
		return err
	}
	if obj.(*riov1.ExternalService).Spec.Service != "" {
		targetStackName, targetServiceName := kv.Split(obj.(*riov1.ExternalService).Spec.Service, "/")
		svc, err := p.serviceCache.Get(namespace.StackNamespace(stack.Namespace, targetStackName), targetServiceName)
		if err != nil {
			return err
		}
		serviceSets, err := serviceset.CollectionServices([]*riov1.Service{svc})
		if err != nil {
			return err
		}
		serviceSet, ok := serviceSets[svc.Name]
		if !ok {
			return err
		}
		populate.VirtualServiceForExternalService(obj.(*riov1.ExternalService), serviceSet, svc, stack, os)
	}
	return nil
}
