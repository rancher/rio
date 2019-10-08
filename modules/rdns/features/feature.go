package features

import (
	"context"

	"github.com/rancher/rio/modules/rdns/controllers/service"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	feature := &features.FeatureController{
		FeatureName: "rdns",
		FeatureSpec: features.FeatureSpec{
			Enabled:     true,
			Description: "Acquire DNS from public Rancher DNS service",
		},
		Controllers: []features.ControllerRegister{
			service.Register,
		},
	}

	return feature.Register()
}
