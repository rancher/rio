package monitoring

import (
	"context"

	v1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	feature := &features.FeatureController{
		FeatureName: "mixer",
		FeatureSpec: v1.FeatureSpec{
			Description: "Istio Mixer telemetry",
			Answers: map[string]string{
				"GRAFANA_USERNAME": "admin",
				"GRAFANA_PASSWORD": "admin",
			},
			Requires: []string{
				"prometheus",
			},
		},
		SystemStacks: []*systemstack.SystemStack{
			systemstack.NewSystemStack(rContext.Apply, rContext.SystemNamespace, rContext.Rio.Rio().V1().Stack(), "istio-telemetry", riov1.StackSpec{
				DisableMesh:               true,
				EnableKubernetesResources: true,
			}),
		},
		FixedAnswers: map[string]string{
			"LB_NAMESPACE": settings.IstioStackName,
		},
	}

	return feature.Register()
}
