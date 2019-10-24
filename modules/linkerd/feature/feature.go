package feature

import (
	"context"

	"github.com/rancher/rio/modules/linkerd/controller/inject"

	"github.com/rancher/rio/modules/linkerd/pkg/injector"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	feature := &features.FeatureController{
		FeatureName: "linkerd",
		FeatureSpec: features.FeatureSpec{
			Description: "linkerd service mesh",
			Enabled:     true,
		},
		SystemStacks: []*stack.SystemStack{},
		Controllers: []features.ControllerRegister{
			inject.Register,
		},
		OnStart: func() error {
			injector.RegisterInjector()
			rContext.Rio.Rio().V1().Service().Enqueue("*", "*")
			return nil
		},
	}
	return feature.Register()
}
