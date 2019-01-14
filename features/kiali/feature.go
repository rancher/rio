package kiali

import (
	"context"
	"fmt"

	"github.com/rancher/rio/pkg/settings"

	v1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"

	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	feature := &features.FeatureController{
		FeatureName: "kiali",
		FeatureSpec: v1.FeatureSpec{
			Description: "Kiali Dashboard",
			Enabled:     false,
			Answers: map[string]string{
				"USERNAME": "admin",
				"PASSWORD": "admin",
			},
		},
		SystemStacks: []*systemstack.SystemStack{
			systemstack.NewSystemStack(rContext.Rio.Stack, "kiali", riov1.StackSpec{
				DisableMesh: true,
			}),
		},
		FixedAnswers: map[string]string{
			"PROMETHEUS_URL": fmt.Sprintf("http://prometheus.%s.svc.cluster.local:9090", settings.PrometheusNamespace),
			"GRAFANA_URL":    fmt.Sprintf("http://grafana.%s.svc.cluster.local:3000", settings.GrafanaNamespace),
		},
	}
	return feature.Register()
}
