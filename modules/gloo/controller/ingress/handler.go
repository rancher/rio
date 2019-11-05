package ingress

import (
	"context"

	rioadminv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	networkingv1beta1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/extensions/v1beta1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := &handler{
		namespace:         rContext.Namespace,
		routers:           rContext.Rio.Rio().V1().Router(),
		services:          rContext.Rio.Rio().V1().Service(),
		publicDomainCache: rContext.Admin.Admin().V1().PublicDomain().Cache(),
	}

	networkingv1beta1controller.RegisterIngressGeneratingHandler(ctx,
		rContext.K8sNetworking.Extensions().V1beta1().Ingress(),
		rContext.Apply.
			WithCacheTypes(rContext.Gateway.Gateway().V1().VirtualService()),
		"",
		"gloo-app",
		h.generate,
		nil)

	return nil
}

type handler struct {
	namespace         string
	routers           riov1controller.RouterController
	services          riov1controller.ServiceController
	publicDomainCache rioadminv1controller.PublicDomainCache
}

func (h handler) generate(obj *v1beta1.Ingress, status v1beta1.IngressStatus) ([]runtime.Object, v1beta1.IngressStatus, error) {
	if obj.Annotations["kubernetes.io/ingress.class"] != "gloo" || obj.Namespace != h.namespace {
		return nil, status, nil
	}

	for _, rule := range obj.Spec.Rules {
		pd, err := h.publicDomainCache.Get(rule.Host)
		if err != nil {
			return nil, status, nil
		}
		if pd.Spec.TargetApp != "" {
			h.services.Enqueue(pd.Spec.TargetNamespace, pd.Spec.TargetApp)
		}
		if pd.Spec.TargetRouter != "" {
			h.routers.Enqueue(pd.Spec.TargetNamespace, pd.Spec.TargetApp)
		}
	}
	return nil, status, nil
}
