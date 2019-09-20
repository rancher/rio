package feature

import (
	"context"

	"github.com/rancher/rio/modules/linkerd/controllers/router"

	"github.com/rancher/rio/modules/linkerd/controllers/app"
	v1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/start"
)

func Register(ctx context.Context, rContext *types.Context) error {
	apply := rContext.Apply.WithCacheTypes(rContext.Rio.Rio().V1().Router())
	feature := &features.FeatureController{
		FeatureName: "linkerd",
		FeatureSpec: v1.FeatureSpec{
			Description: "Linkerd service mesh",
			Enabled:     constants.ServiceMeshMode == constants.ServiceMeshModeLinkerd,
		},
		SystemStacks: []*stack.SystemStack{
			stack.NewSystemStack(apply, "linkerd", "linkerd"),
		},
		Controllers: []features.ControllerRegister{
			app.Register,
			router.Register,
		},
		OnStart: func(feature *v1.Feature) error {
			return start.All(ctx, 5,
				rContext.Apps,
				rContext.Core,
				rContext.SMI)
		},
	}
	return feature.Register()
}
