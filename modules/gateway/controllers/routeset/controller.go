package routeset

import (
	"context"

	"github.com/rancher/rio/modules/gateway/controllers/routeset/populate"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	projectv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/objectset"
	"github.com/rancher/wrangler/pkg/relatedresource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "routing-routeset", rContext.Rio.Rio().V1().Router())
	c.Apply = c.Apply.WithStrictCaching().
		WithCacheTypes(rContext.Networking.Networking().V1alpha3().VirtualService(),
			rContext.Networking.Networking().V1alpha3().DestinationRule(),
			rContext.Networking.Networking().V1alpha3().ServiceEntry())

	r := &routeSetHandler{
		systemNamespace:      rContext.Namespace,
		secretCache:          rContext.Core.Core().V1().Secret().Cache(),
		externalServiceCache: rContext.Rio.Rio().V1().ExternalService().Cache(),
		routesetCache:        rContext.Rio.Rio().V1().Router().Cache(),
		clusterDomainCache:   rContext.Global.Admin().V1().ClusterDomain().Cache(),
	}

	relatedresource.Watch(ctx, "externalservice-routeset", r.resolve,
		rContext.Rio.Rio().V1().Router(), rContext.Rio.Rio().V1().ExternalService())

	c.Populator = r.populate
	return nil
}

func (r routeSetHandler) resolve(namespace, name string, obj runtime.Object) ([]relatedresource.Key, error) {
	switch obj.(type) {
	case *riov1.ExternalService:
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
	secretCache          corev1controller.SecretCache
	externalServiceCache riov1controller.ExternalServiceCache
	routesetCache        riov1controller.RouterCache
	clusterDomainCache   projectv1controller.ClusterDomainCache
}

func (r *routeSetHandler) populate(obj runtime.Object, ns *corev1.Namespace, os *objectset.ObjectSet) error {
	routeSet := obj.(*riov1.Router)
	externalServiceMap := map[string]*riov1.ExternalService{}
	routesetMap := map[string]*riov1.Router{}

	clusterDomain, err := r.clusterDomainCache.Get(r.systemNamespace, constants.ClusterDomainName)
	if err != nil {
		return err
	}
	domain := clusterDomain.Status.ClusterDomain
	if domain == "" {
		return nil
	}

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

	if err := populate.VirtualServices(r.systemNamespace, clusterDomain, obj.(*riov1.Router), externalServiceMap, routesetMap, os); err != nil {
		return err
	}

	return nil
}
