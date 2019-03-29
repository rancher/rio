package build

import (
	"context"

	"github.com/rancher/rio/features/build/buildkit"
	"github.com/rancher/rio/features/build/execution"
	"github.com/rancher/rio/features/build/webhook"
	v1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/rancher/rio/types"
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
