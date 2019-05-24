package publicdomain

import (
	"context"

	"github.com/rancher/rio/modules/service/controllers/publicdomain/populate"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	adminv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	"github.com/rancher/wrangler/pkg/relatedresource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	appPublicdomain    = "app-publicdomain"
	routerPublicdomain = "router-publicdomain"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "stack-public-domain", rContext.Global.Admin().V1().PublicDomain())
	c.Apply = c.Apply.WithCacheTypes(rContext.Extensions.Extensions().V1beta1().Ingress())
	p := populator{
		systemNamespace: rContext.Namespace,
	}

	c.Populator = p.populate
	h := handler{
		services: rContext.Rio.Rio().V1().Service(),
		routers:  rContext.Rio.Rio().V1().Router(),
		apps:     rContext.Rio.Rio().V1().App(),
		domains:  rContext.Global.Admin().V1().PublicDomain().Cache(),
	}

	svcUpdator := riov1controller.UpdateAppOnChange(rContext.Rio.Rio().V1().App().Updater(), h.syncApp)
	rContext.Rio.Rio().V1().App().OnChange(ctx, appPublicdomain, svcUpdator)

	routerUpdator := riov1controller.UpdateRouterOnChange(rContext.Rio.Rio().V1().Router().Updater(), h.syncRouter)
	rContext.Rio.Rio().V1().Router().OnChange(ctx, routerPublicdomain, routerUpdator)

	relatedresource.Watch(ctx, "publicdomain-app", h.resolve,
		rContext.Rio.Rio().V1().App(),
		rContext.Global.Admin().V1().PublicDomain())

	relatedresource.Watch(ctx, "publicdomain-router", h.resolve,
		rContext.Rio.Rio().V1().Router(),
		rContext.Global.Admin().V1().PublicDomain())

	return nil
}

type populator struct {
	systemNamespace string
}

func (p populator) populate(obj runtime.Object, namespace *corev1.Namespace, os *objectset.ObjectSet) error {
	populate.Ingress(p.systemNamespace, obj.(*adminv1.PublicDomain), os)
	return nil
}

type handler struct {
	services riov1controller.ServiceController
	apps     riov1controller.AppController
	routers  riov1controller.RouterController
	domains  adminv1controller.PublicDomainCache
}

func (h handler) resolve(namespace, name string, obj runtime.Object) ([]relatedresource.Key, error) {
	switch obj.(type) {
	case *adminv1.PublicDomain:
		pd := obj.(*adminv1.PublicDomain)
		return []relatedresource.Key{
			{
				Name:      pd.Spec.TargetServiceName,
				Namespace: pd.Namespace,
			},
		}, nil
	}
	return nil, nil
}

func (h handler) syncApp(key string, obj *riov1.App) (*riov1.App, error) {
	if obj == nil || obj.DeletionTimestamp != nil {
		return obj, nil
	}

	pds, err := h.domains.List(obj.Namespace, labels.NewSelector())
	if err != nil {
		return obj, err
	}

	var publicdomains []string
	for _, pd := range pds {
		if pd.Spec.TargetServiceName == obj.Name && pd.DeletionTimestamp == nil {
			publicdomains = append(publicdomains, pd.Spec.DomainName)
		}
	}

	obj.Status.PublicDomains = publicdomains
	return obj, nil
}

func (h handler) syncRouter(key string, obj *riov1.Router) (*riov1.Router, error) {
	if obj == nil || obj.DeletionTimestamp != nil {
		return obj, nil
	}

	pds, err := h.domains.List(obj.Namespace, labels.NewSelector())
	if err != nil {
		return obj, err
	}

	var publicdomains []string
	for _, pd := range pds {
		if pd.Spec.TargetServiceName == obj.Name {
			publicdomains = append(publicdomains, pd.Spec.DomainName)
		}
	}

	obj.Status.PublicDomains = publicdomains
	return obj, nil
}
