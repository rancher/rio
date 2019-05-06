package feature

import (
	"context"

	"github.com/rancher/rio/modules/build/controllers/build"
	"github.com/rancher/rio/modules/build/controllers/gitmodule"

	"github.com/rancher/rio/modules/build/controllers/execution"
	"github.com/rancher/rio/modules/build/controllers/service"
	"github.com/rancher/rio/modules/build/controllers/webhook"
	v1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	apply := rContext.Apply.WithCacheTypes(rContext.Rio.Rio().V1().Service(), rContext.Core.Core().V1().ConfigMap())
	feature := &features.FeatureController{
		FeatureName: "build",
		FeatureSpec: v1.FeatureSpec{
			Description: "Rio Build, from source code to deployment",
			Enabled:     true,
		},
		SystemStacks: []*systemstack.SystemStack{
			systemstack.NewStack(apply, rContext.Namespace, "build", true),
		},
		Controllers: []features.ControllerRegister{
			webhook.Register,
			execution.Register,
			service.Register,
			build.Register,
			gitmodule.Register,
		},
		FixedAnswers: map[string]string{
			"NAMESPACE": rContext.Namespace,
		},
	}
	return feature.Register()
}
