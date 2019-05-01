package feature

import (
	"context"

	"github.com/rancher/rio/modules/autoscale/controller/service"
	v1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	feature := &features.FeatureController{
		FeatureName: "autoscaling",
		FeatureSpec: v1.FeatureSpec{
			Description: "Auto-scaling services based on QPS and requests load",
			Requires: []string{
				"prometheus",
			},
			Enabled: true,
		},
		SystemStacks: []*systemstack.SystemStack{
			systemstack.NewSystemStack(rContext.Apply, rContext.Namespace, "rio-autoscaler"),
		},
		Controllers: []features.ControllerRegister{
			service.Register,
		},
	}
	return feature.Register()
}
