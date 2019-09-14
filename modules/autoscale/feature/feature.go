package feature

import (
	"context"

	"github.com/rancher/rio/modules/autoscale/controller/service"
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
		FeatureName: "autoscaling",
		FeatureSpec: v1.FeatureSpec{
			Description: "Auto-scaling services based on QPS and requests load",
			Enabled:     !constants.DisableAutoscaling,
		},
		SystemStacks: []*stack.SystemStack{
			stack.NewSystemStack(apply, rContext.Namespace, "rio-autoscaler"),
		},
		FixedAnswers: map[string]string{
			"TAG": "v0.1.2",
		},
		Controllers: []features.ControllerRegister{
			service.Register,
		},
		OnStart: func(feature *v1.Feature) error {
			return start.All(ctx, 5,
				rContext.Serving,
				rContext.AutoScale,
			)
		},
	}
	return feature.Register()
}
