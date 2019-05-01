package letsencrypt

import (
	"context"

	"github.com/rancher/rio/modules/system/features/letsencrypt/controllers/issuer"
	"github.com/rancher/rio/modules/system/features/letsencrypt/controllers/publicdomain"
	v1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	feature := &features.FeatureController{
		FeatureName: "letsencrypt",
		FeatureSpec: v1.FeatureSpec{
			Description: "Let's Encrypt",
			Enabled:     true,
			Questions: []riov1.Question{
				{
					Variable:    settings.RioWildcardType,
					Description: "Type of certificates for rio wildcards domain",
					Default:     settings.StagingType,
					Options:     []string{settings.StagingType, settings.ProductionType, settings.SelfSignedType},
					Type:        "enum",
				},
				{
					Variable:    settings.PublicDomainType,
					Description: "Type of certificates for rio public domain",
					Default:     settings.ProductionType,
					Options:     []string{settings.StagingType, settings.ProductionType, settings.SelfSignedType},
					Type:        "enum",
				},
			},
			Answers: map[string]string{
				// todo: registry in build need production server to be fully functional
				settings.RioWildcardType: settings.ProductionType,
				// todo: self-signed only for testing
				settings.PublicDomainType: settings.SelfSignedType,
			},
		},
		FixedAnswers: map[string]string{
			"NAMESPACE": rContext.Namespace,
		},
		SystemStacks: []*systemstack.SystemStack{
			systemstack.NewSystemStack(rContext.Apply, rContext.Namespace, "cert-manager"),
		},
		Controllers: []features.ControllerRegister{
			issuer.Register,
			publicdomain.Register,
		},
	}

	return feature.Register()
}
