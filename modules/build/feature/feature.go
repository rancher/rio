package feature

import (
	"context"

	"github.com/rancher/rio/modules/build/controllers/build"
	"github.com/rancher/rio/modules/build/controllers/gitcommit"
	"github.com/rancher/rio/modules/build/controllers/proxy"
	"github.com/rancher/rio/modules/build/controllers/service"
	stack1 "github.com/rancher/rio/modules/build/controllers/stack"
	v1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/start"
)

func Register(ctx context.Context, rContext *types.Context) error {
	apply := rContext.Apply.WithCacheTypes(rContext.Rio.Rio().V1().Service(), rContext.Core.Core().V1().ConfigMap())
	feature := &features.FeatureController{
		FeatureName: "build",
		FeatureSpec: v1.FeatureSpec{
			Description: "Rio Build, from source code to deployment",
			Enabled:     !constants.DisableBuild,
		},
		SystemStacks: []*stack.SystemStack{
			stack.NewSystemStack(apply, rContext.Namespace, "build"),
		},
		Controllers: []features.ControllerRegister{
			service.Register,
			build.Register,
			gitcommit.Register,
			proxy.Register,
			stack1.Register,
		},
		FixedAnswers: map[string]string{
			"NAMESPACE": rContext.Namespace,
		},
		OnStart: func(feature *v1.Feature) error {
			return start.All(ctx, 5,
				rContext.Global,
				rContext.Build,
				rContext.Core,
				rContext.Webhook,
			)
		},
	}
	return feature.Register()
}
