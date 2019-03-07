package build

import (
	"context"

	"github.com/rancher/rio/features/build/buildkit"

	"github.com/rancher/rio/features/build/execution"

	"github.com/rancher/rio/features/build/webhook"

	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/rancher/rio/types"
	v1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
)

func Register(ctx context.Context, rContext *types.Context) error {
	feature := &features.FeatureController{
		FeatureName: "build",
		FeatureSpec: v1.FeatureSpec{
			Description: "Rio Build, from source code to deployment",
			Enabled:     true,
		},
		SystemStacks: []*systemstack.SystemStack{},
		Controllers: []features.ControllerRegister{
			webhook.Register,
			execution.Register,
			buildkit.Register,
		},
	}
	return feature.Register()
}
