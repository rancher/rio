package features

import (
	"context"

	"github.com/rancher/rio/modules/service/controllers/publicdomain"

	"github.com/rancher/rio/modules/service/controllers/externalservice"

	"github.com/rancher/rio/modules/service/controllers/routeset"
	"github.com/rancher/rio/modules/service/controllers/service"
	"github.com/rancher/rio/modules/service/controllers/serviceset"
	"github.com/rancher/rio/modules/service/controllers/servicestatus"
	projectv1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	feature := &features.FeatureController{
		FeatureName: "stack",
		FeatureSpec: projectv1.FeatureSpec{
			Description: "Rio Stack Based UX - required",
			Enabled:     true,
		},
		Controllers: []features.ControllerRegister{
			externalservice.Register,
			routeset.Register,
			service.Register,
			serviceset.Register,
			servicestatus.Register,
			publicdomain.Register,
		},
	}

	return feature.Register()
}
