package prometheus

import (
	"context"

	v1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	apply := rContext.Apply.WithCacheTypes(rContext.Rio.Rio().V1().Service(), rContext.Core.Core().V1().ConfigMap())
	feature := &features.FeatureController{
		FeatureName: "prometheus",
		FeatureSpec: v1.FeatureSpec{
			Description: "Enable prometheus",
			Enabled:     !constants.DisablePrometheus,
		},
		SystemStacks: []*systemstack.SystemStack{
			systemstack.NewStack(apply, rContext.Namespace, "prometheus", true),
		},
		FixedAnswers: map[string]string{
			"TELEMETRY_NAMESPACE": rContext.Namespace,
			"PILOT_NAME":          constants.IstioStackName,
			"PROMETHEUS_NAME":     constants.Prometheus,
		},
	}
	return feature.Register()
}
