package grafana

import (
	"context"

	v1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/rancher/rio/types"
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
			Enabled: true,
		},
		SystemStacks: []*systemstack.SystemStack{
			systemstack.NewSystemStack(rContext.Apply, rContext.Namespace, "grafana"),
		},
	}
	return feature.Register()
}
