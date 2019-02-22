package prometheus

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
		FeatureName: "prometheus",
		FeatureSpec: v1.FeatureSpec{
			Description: "Enable prometheus",
		},
		SystemStacks: []*systemstack.SystemStack{
			systemstack.NewSystemStack(rContext.Rio.Stack, "prometheus", riov1.StackSpec{}),
		},
		FixedAnswers: map[string]string{
			"TELEMETRY_NAME":  settings.IstioTelemetryNamespace,
			"PILOT_NAME":      settings.IstioPilotNamespace,
			"PROMETHEUS_NAME": settings.PrometheusNamespace,
		},
	}
	return feature.Register()
}
