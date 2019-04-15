package routeset

import (
	"context"
	"fmt"

	"github.com/rancher/wrangler/pkg/generic"

	"github.com/rancher/rio/modules/istio/controllers/routeset/populate"
	"github.com/rancher/rio/modules/istio/pkg/domains"
	projectv1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	projectv1controller "github.com/rancher/rio/pkg/generated/controllers/project.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	"github.com/rancher/wrangler/pkg/relatedresource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	routerDomainUpdate = "router-domain-updater"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "routing-routeset", rContext.Rio.Rio().V1().Router())
	c.Apply = c.Apply.WithStrictCaching().
		WithCacheTypes(rContext.Networking.Networking().V1alpha3().VirtualService(),
			rContext.Networking.Networking().V1alpha3().DestinationRule(),
			rContext.Networking.Networking().V1alpha3().ServiceEntry())

	r := &routeSetHandler{
		systemNamespace:      rContext.Namespace,
		externalServiceCache: rContext.Rio.Rio().V1().ExternalService().Cache(),
		routesetCache:        rContext.Rio.Rio().V1().Router().Cache(),
		clusterDomainCache:   rContext.Global.Project().V1().ClusterDomain().Cache(),
	}

	rContext.Rio.Rio().V1().Router().AddGenericHandler(ctx, routerDomainUpdate, generic.UpdateOnChange(rContext.Rio.Rio().V1().Router().Updater(), r.syncDomain))

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
	externalServiceCache riov1controller.ExternalServiceCache
	routesetCache        riov1controller.RouterCache
	clusterDomainCache   projectv1controller.ClusterDomainCache
}

func (r *routeSetHandler) populate(obj runtime.Object, ns *corev1.Namespace, os *objectset.ObjectSet) error {
	routeSet := obj.(*riov1.Router)
	externalServiceMap := map[string]*riov1.ExternalService{}
	routesetMap := map[string]*riov1.Router{}

	clusterDomain, err := r.clusterDomainCache.Get(r.systemNamespace, settings.ClusterDomainName)
	if err != nil {
		return err
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

	return populate.DestinationRules(obj.(*riov1.Router), os)
}

func (r *routeSetHandler) syncDomain(key string, obj runtime.Object) (runtime.Object, error) {
	if obj == nil {
		return nil, nil
	}

	clusterDomain, err := r.clusterDomainCache.Get(r.systemNamespace, settings.ClusterDomainName)
	if err != nil {
		return obj, err
	}

	updateDomain(obj.(*riov1.Router), clusterDomain)

	return obj, nil
}

func updateDomain(router *riov1.Router, clusterDomain *projectv1.ClusterDomain) {
	protocol := "http"
	if clusterDomain.Status.HTTPSSupported {
		protocol = "https"
	}
	router.Status.Endpoints = []riov1.Endpoint{
		{
			URL: fmt.Sprintf("%s://%s", protocol, domains.GetExternalDomain(router.Name, router.Namespace, clusterDomain.Status.ClusterDomain)),
		},
	}
	for _, pd := range router.Status.PublicDomains {
		router.Status.Endpoints = append(router.Status.Endpoints, riov1.Endpoint{
			URL: fmt.Sprintf("%s://%s", protocol, pd),
		})
	}
}
