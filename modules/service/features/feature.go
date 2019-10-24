package features

import (
	"context"

	"github.com/rancher/rio/modules/service/controllers/app"
	"github.com/rancher/rio/modules/service/controllers/externalservice"
	"github.com/rancher/rio/modules/service/controllers/globalrbac"
	"github.com/rancher/rio/modules/service/controllers/rollout"
	"github.com/rancher/rio/modules/service/controllers/router"
	"github.com/rancher/rio/modules/service/controllers/service"
	"github.com/rancher/rio/modules/service/controllers/servicestatus"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	feature := &features.FeatureController{
		FeatureName: "service",
		FeatureSpec: features.FeatureSpec{
			Enabled:     true,
			Description: "Rio Service Based UX - required",
		},
		Controllers: []features.ControllerRegister{
			app.Register,
			externalservice.Register,
			router.Register,
			service.Register,
			globalrbac.Register,
			servicestatus.Register,
			rollout.Register,
		},
	}

	return feature.Register()
}
