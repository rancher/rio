package rdns

import (
	"context"

	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/wrangler/pkg/start"

	"github.com/rancher/rio/modules/system/features/rdns/controllers/domain"
	projectv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	feature := &features.FeatureController{
		System:      true,
		FeatureName: "rdns",
		FeatureSpec: projectv1.FeatureSpec{
			Description: "Assign cluster a hostname from public Rancher DNS service",
			Enabled:     !constants.DisableRdns,
		},
		Controllers: []features.ControllerRegister{
			domain.Register,
		},
		OnStart: func(feature *projectv1.Feature) error {
			return start.All(ctx, 5)
		},
	}

	return feature.Register()
}
