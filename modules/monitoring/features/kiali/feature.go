package kiali

import (
	"context"

	v1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	apply := rContext.Apply.WithCacheTypes(rContext.Rio.Rio().V1().Service(), rContext.Core.Core().V1().ConfigMap())
	feature := &features.FeatureController{
		FeatureName: "kiali",
		FeatureSpec: v1.FeatureSpec{
			Description: "Kiali Dashboard",
			Enabled:     true,
			Answers: map[string]string{
				"USERNAME": "admin",
				"PASSWORD": "admin",
			},
			Requires: []string{
				"prometheus",
				"grafana",
				"mixer",
			},
		},
		SystemStacks: []*systemstack.SystemStack{
			systemstack.NewStack(apply, rContext.Namespace, "kiali", true),
		},
		FixedAnswers: map[string]string{
			"PROMETHEUS_URL": "http://prometheus:9090",
			"GRAFANA_URL":    "http://grafana:3000",
			"NAMESPACE":      rContext.Namespace,
		},
	}
	return feature.Register()
}
