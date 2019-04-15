package nfs

import (
	"context"

	v1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	feature := &features.FeatureController{
		FeatureName: "nfs",
		FeatureSpec: v1.FeatureSpec{
			Description: "NFS volume driver",
		},
		SystemStacks: []*systemstack.SystemStack{
			//systemstack.NewSystemStack(rContext.Apply, rContext.SystemNamespace, rContext.Rio.Rio().V1().Stack(), "nfs", riov1.StackSpec{
			//	DisableMesh:               true,
			//	EnableKubernetesResources: true,
			//}),
		},
	}

	return feature.Register()
}
