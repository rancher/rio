package ingress

import (
	"context"

	"github.com/rancher/rio/modules/gloo/pkg/vsfactory"
	"github.com/rancher/rio/types"
	networkingv1beta1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/extensions/v1beta1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := &handler{
		namespace: rContext.Namespace,
		vsFactory: vsfactory.New(rContext),
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
	namespace string
	vsFactory *vsfactory.VirtualServiceFactory
}

func (h handler) generate(obj *v1beta1.Ingress, status v1beta1.IngressStatus) ([]runtime.Object, v1beta1.IngressStatus, error) {
	if obj.Annotations["kubernetes.io/ingress.class"] != "gloo" || obj.Namespace != h.namespace {
		return nil, status, nil
	}

	vss, err := h.vsFactory.ForIngress(obj)
	return vss, status, err
}
