package grafana

import (
	"context"

	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/rancher/rio/types"
	v1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

func Register(ctx context.Context, rContext *types.Context) error {
	feature := &features.FeatureController{
		FeatureName: "grafana",
		FeatureSpec: v1.FeatureSpec{
			Description: "Grafana Dashboard",
			Requires: []string{
				"prometheus",
				"mixer",
			},
		},
		SystemStacks: []*systemstack.SystemStack{
			systemstack.NewSystemStack(rContext.Rio.Stack, "grafana", riov1.StackSpec{}),
		},
		FixedAnswers: map[string]string{
			"PROMETHEUS_NAMESPACE": settings.PrometheusNamespace,
		},
	}
	return feature.Register()
}
