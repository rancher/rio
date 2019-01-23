package rdns

import (
	"context"

	"github.com/rancher/rio/features/rdns/controllers/domain"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/types"
	projectv1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
)

func Register(ctx context.Context, rContext *types.Context) error {
	feature := &features.FeatureController{
		FeatureName: "rdns",
		FeatureSpec: projectv1.FeatureSpec{
			Description: "Assign cluster a hostname from public Rancher DNS service",
			Enabled:     true,
			Requires:    []string{"letsencrypt"},
		},
		Controllers: []features.ControllerRegister{
			domain.Register,
		},
	}

	return feature.Register()
}
