package localstorage

import (
	"context"

	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

func Register(ctx context.Context, rContext *types.Context) error {
	feature := &features.FeatureController{
		FeatureName: "localstorage",
		FeatureSpec: v1.FeatureSpec{
			Description: "Local storage driver",
		},
		SystemStacks: []*systemstack.SystemStack{
			systemstack.NewSystemStack(rContext.Rio.Stack, "localstorage", riov1.StackSpec{
				DisableMesh:               true,
				EnableKubernetesResources: true,
			}),
		},
	}

	return feature.Register()
}
