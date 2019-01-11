package stack

import (
	"context"

	"github.com/rancher/rio/features/stack/controllers/stackns"

	"github.com/rancher/rio/features/stack/controllers/config"
	"github.com/rancher/rio/features/stack/controllers/externalservice"
	"github.com/rancher/rio/features/stack/controllers/pvc"
	"github.com/rancher/rio/features/stack/controllers/routeset"
	"github.com/rancher/rio/features/stack/controllers/service"
	"github.com/rancher/rio/features/stack/controllers/servicestatus"
	"github.com/rancher/rio/features/stack/controllers/stack"
	"github.com/rancher/rio/features/stack/controllers/volume"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/types"
	projectv1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
)

func Register(ctx context.Context, rContext *types.Context) error {
	feature := &features.FeatureController{
		FeatureName: "stack",
		FeatureSpec: projectv1.FeatureSpec{
			Description: "Rio Stack Based UX - required",
			Enabled:     true,
		},
		Controllers: []features.ControllerRegister{
			config.Register,
			externalservice.Register,
			pvc.Register,
			routeset.Register,
			service.Register,
			servicestatus.Register,
			stackns.Register,
			stack.Register,
			volume.Register,
		},
	}

	return feature.Register()
}
