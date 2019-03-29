package routeset

import (
	"context"

	"github.com/rancher/rio/features/routing/controllers/routeset/populate"
	"github.com/rancher/rio/features/routing/controllers/util"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v12 "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	"github.com/rancher/wrangler/pkg/relatedresource"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "routing-routeset", rContext.Rio.Rio().V1().Router())
	c.Apply.WithStrictCaching().
		WithCacheTypes(rContext.Networking.Networking().V1alpha3().VirtualService(),
			rContext.Networking.Networking().V1alpha3().DestinationRule(),
			rContext.Networking.Networking().V1alpha3().ServiceEntry())

	r := &routeSetHandler{
		systemNamespace:      rContext.SystemNamespace,
		externalServiceCache: rContext.Rio.Rio().V1().ExternalService().Cache(),
		routesetCache:        rContext.Rio.Rio().V1().Router().Cache(),
	}

	relatedresource.Watch(ctx, "externalservice-routeset", r.resolve,
		rContext.Rio.Rio().V1().Router(), rContext.Rio.Rio().V1().ExternalService())

	c.Populator = r.populate
	return nil
}

func (r routeSetHandler) resolve(namespace, name string, obj runtime.Object) ([]relatedresource.Key, error) {
	switch obj.(type) {
	case *v1.ExternalService:
		routesets, err := r.routesetCache.List(namespace, labels.Everything())
		if err != nil {
			return nil, err
		}
		var result []relatedresource.Key
		for _, r := range routesets {
			result = append(result, relatedresource.NewKey(r.Namespace, r.Name))
		}
		return result, nil
	}
	return nil, nil
}

type routeSetHandler struct {
	systemNamespace      string
	externalServiceCache v12.ExternalServiceCache
	routesetCache        v12.RouterCache
}

func (r *routeSetHandler) populate(obj runtime.Object, stack *v1.Stack, os *objectset.ObjectSet) error {
	if stack != nil && stack.Spec.DisableMesh {
		return nil
	}

	if err := util.WaitForClusterDomain(); err != nil {
		return err
	}

	routeSet := obj.(*v1.Router)
	externalServiceMap := map[string]*v1.ExternalService{}
	routesetMap := map[string]*v1.Router{}

	ess, err := r.externalServiceCache.List(routeSet.Namespace, labels.Everything())
	if err != nil {
		return err
	}
	for _, es := range ess {
		externalServiceMap[es.Name] = es
	}

	routesets, err := r.routesetCache.List(routeSet.Namespace, labels.Everything())
	if err != nil {
		return err
	}
	for _, rs := range routesets {
		routesetMap[rs.Name] = rs
	}

	if err := populate.VirtualServices(r.systemNamespace, stack, obj.(*v1.Router), externalServiceMap, routesetMap, os); err != nil {
		return err
	}

	return populate.DestinationRules(stack, obj.(*v1.Router), os)
}
