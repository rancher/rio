package kiali

import (
	"context"
	"encoding/base64"
	"fmt"

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
				"USERNAME": base64.StdEncoding.EncodeToString([]byte("admin")),
				"PASSWORD": base64.StdEncoding.EncodeToString([]byte("admin")),
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
			"PROMETHEUS_URL": fmt.Sprintf("http://prometheus.%s:9090", rContext.Namespace),
			"GRAFANA_URL":    fmt.Sprintf("http://grafana.%s:3000", rContext.Namespace),
			"NAMESPACE":      rContext.Namespace,
		},
	}
	return feature.Register()
}
