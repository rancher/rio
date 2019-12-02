package feature

import (
	"context"

	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	apply := rContext.Apply.WithCacheTypes(rContext.Rio.Rio().V1().Service())
	feature := &features.FeatureController{
		FeatureName: "dashboard",
		FeatureSpec: features.FeatureSpec{
			Description: "Rio UI",
		},
		SystemStacks: []*stack.SystemStack{
			stack.NewSystemStack(apply, rContext.Admin.Admin().V1().SystemStack(), rContext.Namespace, "dashboard"),
		},
		FixedAnswers: map[string]string{
			"NAMESPACE": rContext.Namespace,
		},
	}
	return feature.Register()
}
