package features

import (
	"context"

	"github.com/rancher/rio/modules/letsencrypt/controllers/clusterdomain"
	"github.com/rancher/rio/modules/letsencrypt/controllers/issuer"
	"github.com/rancher/rio/modules/letsencrypt/controllers/publicdomain"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	apply := rContext.Apply.WithCacheTypes(rContext.Rio.Rio().V1().Service(), rContext.Core.Core().V1().ConfigMap())
	feature := &features.FeatureController{
		FeatureName: "letsencrypt",
		FeatureSpec: features.FeatureSpec{
			Enabled:     true,
			Description: "Let's Encrypt",
		},
		FixedAnswers: map[string]string{
			"TAG":       "v0.11.0-rio.1",
			"NAMESPACE": rContext.Namespace,
		},
		SystemStacks: []*stack.SystemStack{
			stack.NewSystemStack(apply, rContext.Admin.Admin().V1().SystemStack(), rContext.Namespace, "cert-manager"),
		},
		Controllers: []features.ControllerRegister{
			issuer.Register,
			publicdomain.Register,
			clusterdomain.Register,
		},
	}

	return feature.Register()
}
