package grafana

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
		FeatureName: "grafana",
		FeatureSpec: v1.FeatureSpec{
			Description: "Grafana Dashboard",
			Requires: []string{
				"prometheus",
				"mixer",
			},
			Enabled: !constants.DisableGrafana,
		},
		SystemStacks: []*stack.SystemStack{
			stack.NewSystemStack(apply, rContext.Namespace, "grafana"),
		},
	}
	return feature.Register()
}
