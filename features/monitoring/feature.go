package monitoring

import (
	"context"

	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

func Register(ctx context.Context, rContext *types.Context) error {
	feature := &features.FeatureController{
		FeatureName: "monitoring",
		FeatureSpec: v1.FeatureSpec{
			Description: "Monitoring and Telemetry",
		},
		SystemStacks: []*systemstack.SystemStack{
			systemstack.NewSystemStack(rContext.Rio.Stack, "istio-telemetry", riov1.StackSpec{
				DisableMesh: true,
				Answers: map[string]string{
					"LB_NAMESPACE": settings.IstioExternalLBNamespace,
				},
				EnableKubernetesResources: true,
			}),
		},
	}

	return feature.Register()
}
