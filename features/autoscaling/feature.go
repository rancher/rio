package autoscaling

import (
	"context"

	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/rancher/rio/types"
	v1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

func Register(ctx context.Context, rContext *types.Context) error {
	feature := &features.FeatureController{
		FeatureName: "autoscaling",
		FeatureSpec: v1.FeatureSpec{
			Description: "Auto-scaling services based on QPS and requests load",
			Requires: []string{
				"prometheus",
			},
		},
		SystemStacks: []*systemstack.SystemStack{
			systemstack.NewSystemStack(rContext.Rio.Stack, "rio-autoscaler", riov1.StackSpec{}),
		},
	}
	return feature.Register()
}
