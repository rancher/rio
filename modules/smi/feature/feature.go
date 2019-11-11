package features

import (
	"context"

	"github.com/rancher/rio/modules/smi/controllers/app"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	feature := &features.FeatureController{
		FeatureName: "smi",
		FeatureSpec: features.FeatureSpec{
			Description: "Program SMI rules for services",
			Enabled:     true,
		},
		Controllers: []features.ControllerRegister{
			app.Register,
		},
	}

	return feature.Register()
}
