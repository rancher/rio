package feature

import (
	"context"

	"github.com/rancher/rio/pkg/arch"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	feature := &features.FeatureController{
		FeatureName: "smi-metrics",
		FeatureSpec: features.FeatureSpec{

			Enabled:     arch.IsAmd64(),
			Description: "APIService for service-mesh metrics",
			//TODO enable different adapters based enabled mesh Linkerd/Istio
			Questions: nil,
			Answers:   nil,
			Requires:  nil,
		},
		Controllers: nil,
		OnStop:      nil,
		OnStart:     nil,
		SystemStacks: []*stack.SystemStack{
			stack.NewSystemStack(rContext.Apply, rContext.Admin.Admin().V1().SystemStack(), rContext.Namespace, "smi-metrics"),
		},
	}

	return feature.Register()
}
