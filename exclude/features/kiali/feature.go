package kiali

import (
	"context"

	v1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
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
			Requires: []string{
				"prometheus",
				"grafana",
				"mixer",
			},
		},
		SystemStacks: []*systemstack.SystemStack{
			systemstack.NewSystemStack(rContext.Apply, rContext.SystemNamespace, rContext.Rio.Rio().V1().Stack(), "kiali", riov1.StackSpec{}),
		},
		FixedAnswers: map[string]string{
			"PROMETHEUS_URL": "http://prometheus",
			"GRAFANA_URL":    "http://grafana",
		},
	}
	return feature.Register()
}
