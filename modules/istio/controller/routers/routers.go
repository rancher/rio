package routers

import (
	"context"

	"github.com/rancher/rio/modules/gloo/pkg/vsfactory"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContexts *types.Context) error {
	h := &handler{
		vsFactory: vsfactory.New(rContexts),
	}

	riov1controller.RegisterRouterGeneratingHandler(ctx,
		rContexts.Rio.Rio().V1().Router(),
		rContexts.Apply.
			WithCacheTypes(rContexts.Istio.Networking().V1alpha3().VirtualService()),
		"GatewayConfigured",
		"istio-router",
		h.generate,
		nil)

	return nil
}

type handler struct {
	vsFactory *vsfactory.VirtualServiceFactory
}

func (h *handler) generate(router *riov1.Router, status riov1.RouterStatus) ([]runtime.Object, riov1.RouterStatus, error) {
	vs, err := h.vsFactory.ForIstioRouter(router)
	if err != nil {
		return nil, status, err
	}

	return []runtime.Object{vs}, status, nil
}
