package features

import (
	"context"

	"github.com/rancher/rio/modules/gloo/controller/app"
	"github.com/rancher/rio/modules/gloo/controller/config"
	"github.com/rancher/rio/modules/gloo/controller/router"
	"github.com/rancher/rio/modules/gloo/controller/service"
	"github.com/rancher/rio/modules/gloo/pkg/injector"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	injector.RegisterInjector()
	apply := rContext.Apply.WithInjectorName("gloo-mesh")
	feature := &features.FeatureController{
		FeatureName: "gloo",
		FeatureSpec: features.FeatureSpec{
			Enabled:     true,
			Description: "Run Gloo API gateway",
		},
		FixedAnswers: map[string]string{
			"NAMESPACE": rContext.Namespace,
		},
		SystemStacks: []*stack.SystemStack{
			stack.NewSystemStack(apply, rContext.Namespace, "gloo"),
		},
		Controllers: []features.ControllerRegister{
			app.Register,
			router.Register,
			config.Register,
			service.Register,
		},
	}

	return feature.Register()
}
