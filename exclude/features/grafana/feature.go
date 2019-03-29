package grafana

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
		FeatureName: "grafana",
		FeatureSpec: v1.FeatureSpec{
			Description: "Grafana Dashboard",
			Requires: []string{
				"prometheus",
				"mixer",
			},
		},
		SystemStacks: []*systemstack.SystemStack{
			systemstack.NewSystemStack(rContext.Apply, rContext.SystemNamespace, rContext.Rio.Rio().V1().Stack(), "grafana", riov1.StackSpec{}),
		},
	}
	return feature.Register()
}
