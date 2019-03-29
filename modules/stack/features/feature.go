package features

import (
	"context"

	"github.com/rancher/rio/modules/stack/controllers/config"
	"github.com/rancher/rio/modules/stack/controllers/pvc"
	"github.com/rancher/rio/modules/stack/controllers/routeset"
	"github.com/rancher/rio/modules/stack/controllers/service"
	"github.com/rancher/rio/modules/stack/controllers/servicestatus"
	"github.com/rancher/rio/modules/stack/controllers/stack"
	"github.com/rancher/rio/modules/stack/controllers/stackns"
	"github.com/rancher/rio/modules/stack/controllers/volume"
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
			config.Register,
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
