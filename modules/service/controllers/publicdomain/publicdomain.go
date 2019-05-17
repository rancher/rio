package publicdomain

import (
	"context"

	"k8s.io/apimachinery/pkg/labels"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/relatedresource"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	servicePublicdomain = "service-publicdomain"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		services: rContext.Rio.Rio().V1().Service(),
		routers:  rContext.Rio.Rio().V1().Router(),
		apps:     rContext.Rio.Rio().V1().App(),
		domains:  rContext.Rio.Rio().V1().PublicDomain().Cache(),
	}

	svcUpdator := riov1controller.UpdateAppOnChange(rContext.Rio.Rio().V1().App().Updater(), h.syncApp)
	rContext.Rio.Rio().V1().App().OnChange(ctx, servicePublicdomain, svcUpdator)

	routerUpdator := riov1controller.UpdateRouterOnChange(rContext.Rio.Rio().V1().Router().Updater(), h.syncRouter)
	rContext.Rio.Rio().V1().Router().OnChange(ctx, servicePublicdomain, routerUpdator)

	relatedresource.Watch(ctx, "publicdomain-app", h.resolve,
		rContext.Rio.Rio().V1().App(),
		rContext.Rio.Rio().V1().PublicDomain())

	relatedresource.Watch(ctx, "publicdomain-router", h.resolve,
		rContext.Rio.Rio().V1().Router(),
		rContext.Rio.Rio().V1().PublicDomain())

	return nil
}

type handler struct {
	services riov1controller.ServiceController
	apps     riov1controller.AppController
	routers  riov1controller.RouterController
	domains  riov1controller.PublicDomainCache
}

func (h handler) resolve(namespace, name string, obj runtime.Object) ([]relatedresource.Key, error) {
	switch obj.(type) {
	case *riov1.PublicDomain:
		pd := obj.(*riov1.PublicDomain)
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
		if pd.Spec.TargetServiceName == obj.Name {
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
