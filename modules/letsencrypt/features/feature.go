package features

import (
	"context"

	"github.com/rancher/rio/modules/letsencrypt/controllers/account"
	"github.com/rancher/rio/modules/letsencrypt/controllers/certificate"
	"github.com/rancher/rio/modules/letsencrypt/controllers/clusterdomain"
	"github.com/rancher/rio/modules/letsencrypt/controllers/publicdomain"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	feature := &features.FeatureController{
		FeatureName: "letsencrypt",
		FeatureSpec: features.FeatureSpec{
			Enabled:     true,
			Description: "Let's Encrypt",
		},
		SystemStacks: []*stack.SystemStack{},
		Controllers: []features.ControllerRegister{
			account.Register,
			publicdomain.Register,
			clusterdomain.Register,
			certificate.Register,
		},
	}

	return feature.Register()
}
