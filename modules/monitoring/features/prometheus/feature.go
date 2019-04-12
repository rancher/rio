package prometheus

import (
	"context"

	v1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	feature := &features.FeatureController{
		FeatureName: "prometheus",
		FeatureSpec: v1.FeatureSpec{
			Description: "Enable prometheus",
		},
		SystemStacks: []*systemstack.SystemStack{
			//systemstack.NewSystemStack(rContext.Apply, rContext.SystemNamespace, rContext.Rio.Rio().V1().Stack(), "prometheus", riov1.StackSpec{}),
		},
		FixedAnswers: map[string]string{
			"TELEMETRY_NAME":  settings.IstioTelemetry,
			"PILOT_NAME":      settings.IstioStackName,
			"PROMETHEUS_NAME": settings.Prometheus,
		},
	}
	return feature.Register()
}
