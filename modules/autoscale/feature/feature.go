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
		FeatureName: "autoscaling",
		FeatureSpec: features.FeatureSpec{
			Description: "Auto-scaling services based on in-flight requests",
			Enabled:     true,
		},
		SystemStacks: []*stack.SystemStack{
			stack.NewSystemStack(apply, rContext.Namespace, "rio-autoscaler"),
		},
		FixedAnswers: map[string]string{
			"NAMESPACE": rContext.Namespace,
		},
	}
	return feature.Register()
}
