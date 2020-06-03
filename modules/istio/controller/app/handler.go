package app

import (
	"context"
	"fmt"

	"github.com/rancher/rio/modules/gloo/pkg/vsfactory"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/indexes"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		serviceCache: rContext.Rio.Rio().V1().Service().Cache(),
		vsFactory:    vsfactory.New(rContext),
	}
	corev1controller.RegisterServiceGeneratingHandler(ctx,
		rContext.Core.Core().V1().Service(),
		rContext.Apply.
			WithCacheTypes(rContext.Istio.Networking().V1alpha3().VirtualService(),
				rContext.Istio.Networking().V1alpha3().DestinationRule()),
		"",
		"istio-app",
		h.generate,
		nil)

	return nil
}

type handler struct {
	serviceCache riov1controller.ServiceCache
	vsFactory    *vsfactory.VirtualServiceFactory
}

func (h handler) generate(svc *v1.Service, status v1.ServiceStatus) ([]runtime.Object, v1.ServiceStatus, error) {
	app := svc.Labels["rio.cattle.io/app"]
	if app == "" {
		return nil, status, nil
	}

	svcs, err := h.serviceCache.GetByIndex(indexes.ServiceByApp, fmt.Sprintf("%s/%s", svc.Namespace, app))
	if err != nil {
		return nil, status, err
	}

	if len(svcs) == 0 {
		return nil, status, nil
	}

	vss, dest, err := h.vsFactory.ForAppIstio(svc.Namespace, app, svcs)
	if err != nil {
		return nil, status, err
	}

	return []runtime.Object{vss, dest}, status, nil
}
