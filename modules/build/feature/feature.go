package feature

import (
	"context"

	"github.com/rancher/rio/modules/build/controllers/build"
	"github.com/rancher/rio/modules/build/controllers/gitcommit"
	"github.com/rancher/rio/modules/build/controllers/proxy"
	"github.com/rancher/rio/modules/build/controllers/service"
	stack1 "github.com/rancher/rio/modules/build/controllers/stack"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	apply := rContext.Apply.WithCacheTypes(rContext.Rio.Rio().V1().Service(), rContext.Core.Core().V1().ConfigMap())
	feature := &features.FeatureController{
		FeatureName: "build",
		FeatureSpec: features.FeatureSpec{
			Description: "Rio Build, from source code to deployment",
			Enabled:     true,
		},
		SystemStacks: []*stack.SystemStack{
			stack.NewSystemStack(apply, rContext.Admin.Admin().V1().SystemStack(), rContext.Namespace, "build"),
			stack.NewSystemStack(rContext.Apply, rContext.Admin.Admin().V1().SystemStack(), rContext.Namespace, "tekton"),
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
	}
	return feature.Register()
}
