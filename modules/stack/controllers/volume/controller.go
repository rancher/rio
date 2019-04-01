package volume

import (
	"context"

	"github.com/rancher/rio/modules/stack/controllers/volume/populate"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	sc := stackobject.NewGeneratingController(ctx, rContext, "volume", rContext.Rio.Rio().V1().Volume())
	sc.Apply = sc.Apply.WithCacheTypes(rContext.Core.Core().V1().PersistentVolumeClaim())

	sc.Populator = func(obj runtime.Object, stack *riov1.Stack, os *objectset.ObjectSet) error {
		return populate.Volume(obj.(*riov1.Volume), os)
	}

	return nil
}
