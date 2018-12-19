package routeset

import (
	"context"

	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/features/routing/controllers/routeset/populate"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "routing-routeset", rContext.Rio.RouteSet)
	c.Processor.Client(rContext.Networking.VirtualService)

	r := &routeSetHandler{
		externalServiceCache: rContext.Rio.ExternalService.Cache(),
	}

	c.Populator = r.populate

	return nil
}

type routeSetHandler struct {
	externalServiceCache v1.ExternalServiceClientCache
}

func (r *routeSetHandler) populate(obj runtime.Object, stack *v1.Stack, os *objectset.ObjectSet) error {
	routeSet := obj.(*v1.RouteSet)
	externalServiceMap := map[string]*v1.ExternalService{}

	// TODO: What if an external service changes, do we watch that?
	ess, err := r.externalServiceCache.List(routeSet.Namespace, labels.Everything())
	if err != nil {
		return err
	}

	for _, es := range ess {
		externalServiceMap[es.Name] = es
	}

	return populate.VirtualServices(stack, obj.(*v1.RouteSet), externalServiceMap, os)
}
