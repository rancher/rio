package letsencrypt

import (
	"context"

	"github.com/rancher/rio/modules/system/features/letsencrypt/controllers/issuer"
	"github.com/rancher/rio/modules/system/features/letsencrypt/controllers/publicdomain"
	v1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/start"
)

func Register(ctx context.Context, rContext *types.Context) error {
	apply := rContext.Apply.WithCacheTypes(rContext.Rio.Rio().V1().Service(), rContext.Core.Core().V1().ConfigMap())
	feature := &features.FeatureController{
		FeatureName: "letsencrypt",
		FeatureSpec: v1.FeatureSpec{
			Description: "Let's Encrypt",
			Enabled:     !constants.DisableLetsencrypt,
			Questions: []riov1.Question{
				{
					Variable:    constants.RioWildcardType,
					Description: "Type of certificates for rio wildcards domain",
					Default:     constants.StagingType,
					Options:     []string{constants.StagingType, constants.ProductionType, constants.SelfSignedType},
					Type:        "enum",
				},
				{
					Variable:    constants.PublicDomainType,
					Description: "Type of certificates for rio public domain",
					Default:     constants.ProductionType,
					Options:     []string{constants.StagingType, constants.ProductionType, constants.SelfSignedType},
					Type:        "enum",
				},
			},
			Answers: map[string]string{
				constants.RioWildcardType:  constants.ProductionType,
				constants.PublicDomainType: constants.ProductionType,
			},
		},
		FixedAnswers: map[string]string{
			"TAG":       "v0.7.2-rio.1",
			"NAMESPACE": rContext.Namespace,
		},
		SystemStacks: []*stack.SystemStack{
			stack.NewSystemStack(apply, rContext.Namespace, "cert-manager"),
		},
		Controllers: []features.ControllerRegister{
			issuer.Register,
			publicdomain.Register,
		},
		OnStart: func(feature *v1.Feature) error {
			return start.All(ctx, 5,
				rContext.CertManager,
			)
		},
	}

	return feature.Register()
}
