package volume

import (
	"context"

	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/features/stack/controllers/volume/populate"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	sc := stackobject.NewGeneratingController(ctx, rContext, "volume", rContext.Rio.Volume)
	sc.Processor.Client(rContext.Core.PersistentVolumeClaim)

	sc.Populator = func(obj runtime.Object, stack *riov1.Stack, os *objectset.ObjectSet) error {
		return populate.Volume(stack, obj.(*riov1.Volume), os)
	}

	return nil
}
