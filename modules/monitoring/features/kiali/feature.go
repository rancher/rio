package kiali

import (
	"context"

	"github.com/rancher/rio/pkg/constants"

	v1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	apply := rContext.Apply.WithCacheTypes(rContext.Rio.Rio().V1().Service(), rContext.Core.Core().V1().ConfigMap())
	feature := &features.FeatureController{
		FeatureName: "kiali",
		FeatureSpec: v1.FeatureSpec{
			Description: "Kiali Dashboard",
			Enabled:     !constants.DisableKiali,
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
		SystemStacks: []*stack.SystemStack{
			stack.NewSystemStack(apply, rContext.Namespace, "kiali"),
		},
		FixedAnswers: map[string]string{
			"PROMETHEUS_URL": "http://prometheus:9090",
			"GRAFANA_URL":    "http://grafana:3000",
			"NAMESPACE":      rContext.Namespace,
		},
	}
	return feature.Register()
}
