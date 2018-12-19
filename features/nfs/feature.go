package nfs

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
		FeatureName: "nfs",
		FeatureSpec: v1.FeatureSpec{
			Description: "Enable nfs volume feature",
		},
		SystemStacks: []*systemstack.SystemStack{
			systemstack.NewSystemStack(rContext.Rio.Stack, "nfs", riov1.StackSpec{
				DisableMesh:               true,
				EnableKubernetesResources: true,
			}),
		},
	}

	return feature.Register()
}
