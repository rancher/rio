package letsencrypt

import (
	"context"

	"github.com/rancher/rio/features/letsencrypt/controllers/issuer"
	"github.com/rancher/rio/features/letsencrypt/controllers/publicdomain"
	"github.com/rancher/rio/features/letsencrypt/controllers/secrets"
	"github.com/rancher/rio/features/letsencrypt/controllers/service"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"github.com/rancher/types/apis/management.cattle.io/v3"
)

func Register(ctx context.Context, rContext *types.Context) error {
	feature := &features.FeatureController{
		FeatureName: "letsencrypt",
		FeatureSpec: v1.FeatureSpec{
			Description: "Let's Encrypt",
			Enabled:     true,
			Questions: []v3.Question{
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
				settings.RioWildcardType:  settings.StagingType,
				settings.PublicDomainType: settings.ProductionType,
			},
		},
		FixedAnswers: map[string]string{
			settings.CertManagerImageType: settings.CertManagerImage.Get(),
		},
		SystemStacks: []*systemstack.SystemStack{
			systemstack.NewSystemStack(rContext.Rio.Stack, "cert-manager", riov1.StackSpec{
				DisableMesh:               true,
				EnableKubernetesResources: true,
			}),
		},
		Controllers: []features.ControllerRegister{
			issuer.Register,
			publicdomain.Register,
			secrets.Register,
			service.Register,
		},
	}

	return feature.Register()
}
