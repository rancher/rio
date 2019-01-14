package routeset

import (
	"context"

	"github.com/rancher/norman/pkg/changeset"
	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/features/routing/controllers/routeset/populate"
	"github.com/rancher/rio/features/routing/controllers/util"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	v1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "routing-routeset", rContext.Rio.RouteSet)
	c.Processor.Client(rContext.Networking.VirtualService)

	r := &routeSetHandler{
		externalServiceCache: rContext.Rio.ExternalService.Cache(),
		routesetCache:        rContext.Rio.RouteSet.Cache(),
	}

	changeset.Watch(ctx, "externalservice-routeset", r.resolve, rContext.Rio.RouteSet, rContext.Rio.ExternalService)

	c.Populator = r.populate

	return nil
}

func (r routeSetHandler) resolve(namespace, name string, obj runtime.Object) ([]changeset.Key, error) {
	switch obj.(type) {
	case *v1.ExternalService:
		routesets, err := r.routesetCache.List(namespace, labels.Everything())
		if err != nil {
			return nil, err
		}
		var result []changeset.Key
		for _, r := range routesets {
			result = append(result, changeset.Key{
				Namespace: r.Namespace,
				Name:      r.Name,
			})
		}
		return result, nil
	}
	return nil, nil
}

type routeSetHandler struct {
	externalServiceCache v1.ExternalServiceClientCache
	routesetCache        v1.RouteSetClientCache
}

func (r *routeSetHandler) populate(obj runtime.Object, stack *v1.Stack, os *objectset.ObjectSet) error {
	if stack != nil && stack.Spec.DisableMesh {
		return nil
	}

	if err := util.WaitForClusterDomain(); err != nil {
		return err
	}

	routeSet := obj.(*v1.RouteSet)
	externalServiceMap := map[string]*v1.ExternalService{}

	ess, err := r.externalServiceCache.List(routeSet.Namespace, labels.Everything())
	if err != nil {
		return err
	}

	for _, es := range ess {
		externalServiceMap[es.Name] = es
	}

	return populate.VirtualServices(stack, obj.(*v1.RouteSet), externalServiceMap, os)
}
