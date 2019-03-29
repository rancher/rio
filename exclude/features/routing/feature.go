package routing

import (
	"context"

	"github.com/rancher/rio/features/routing/controllers/externalservice"
	"github.com/rancher/rio/features/routing/controllers/istio"
	"github.com/rancher/rio/features/routing/controllers/publicdomain"
	"github.com/rancher/rio/features/routing/controllers/routeset"
	"github.com/rancher/rio/features/routing/controllers/service"
	projectv1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	feature := &features.FeatureController{
		FeatureName: "routing",
		FeatureSpec: projectv1.FeatureSpec{
			Description: "Service routing",
			Enabled:     true,
		},
		SystemStacks: []*systemstack.SystemStack{
			systemstack.NewSystemStack(rContext.Apply, rContext.SystemNamespace, rContext.Rio.Rio().V1().Stack(), "istio-crd", v1.StackSpec{
				DisableMesh:               true,
				EnableKubernetesResources: true,
			}),
		},
		Controllers: []features.ControllerRegister{
			externalservice.Register,
			istio.Register,
			publicdomain.Register,
			routeset.Register,
			service.Register,
		},
	}

	return feature.Register()
}
