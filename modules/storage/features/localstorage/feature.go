package localstorage

import (
	"context"

	v1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	apply := rContext.Apply.WithCacheTypes(rContext.Rio.Rio().V1().Service(), rContext.Core.Core().V1().ConfigMap())
	feature := &features.FeatureController{
		FeatureName: "localstorage",
		FeatureSpec: v1.FeatureSpec{
			Description: "Local storage driver",
			Enabled:     true,
		},
		SystemStacks: []*systemstack.SystemStack{
			systemstack.NewStack(apply, rContext.Namespace, "localstorage", true),
		},
	}

	return feature.Register()
}
