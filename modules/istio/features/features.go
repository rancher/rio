package features

import (
	"context"

	"github.com/rancher/rio/modules/istio/controller/app"
	"github.com/rancher/rio/modules/istio/controller/gateway"
	"github.com/rancher/rio/modules/istio/controller/ingress"
	"github.com/rancher/rio/modules/istio/controller/routers"
	"github.com/rancher/rio/modules/istio/controller/service"
	"github.com/rancher/rio/modules/linkerd/pkg/injector"
	"github.com/rancher/rio/pkg/arch"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	feature := &features.FeatureController{
		FeatureName: "istio",
		FeatureSpec: features.FeatureSpec{
			Enabled:     arch.IsAmd64(),
			Description: "Istio service mesh",
		},
		Controllers: []features.ControllerRegister{
			app.Register,
			routers.Register,
			service.Register,
			gateway.Register,
			ingress.Register,
		},
		OnStart: func() error {
			injector.RegisterInjector()
			rContext.Rio.Rio().V1().Service().Enqueue("*", "*")
			return nil
		},
	}

	return feature.Register()
}
