package feature

import (
	"context"

	"github.com/rancher/rio/modules/linkerd/controller/inject"

	"github.com/rancher/rio/modules/linkerd/pkg/injector"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Register(ctx context.Context, rContext *types.Context) error {
	apply := rContext.Apply.WithCacheTypes(rContext.Rio.Rio().V1().Router())
	feature := &features.FeatureController{
		FeatureName: "linkerd",
		FeatureSpec: features.FeatureSpec{
			Description: "linkerd service mesh",
			Enabled:     true,
		},
		SystemStacks: []*stack.SystemStack{
			stack.NewSystemStack(apply, rContext.Admin.Admin().V1().SystemStack(), "linkerd", "linkerd"),
		},
		Controllers: []features.ControllerRegister{
			inject.Register,
		},
		OnStart: func() error {
			injector.RegisterInjector()
			rContext.Rio.Rio().V1().Service().Enqueue("*", "*")

			settings, err := rContext.Gloo.Gloo().V1().Settings().Get(rContext.Namespace, "default", metav1.GetOptions{})
			if err != nil {
				return err
			}
			settings.Spec.Linkerd = true
			_, err = rContext.Gloo.Gloo().V1().Settings().Update(settings)
			return err
		},
	}
	return feature.Register()
}
