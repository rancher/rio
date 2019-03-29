package routeset

import (
	"context"

	"github.com/rancher/rio/modules/stack/controllers/routeset/populate"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "stack-route-set", rContext.Rio.Rio().V1().Router())
	c.Apply = c.Apply.WithCacheTypes(rContext.Core.Core().V1().Service(), rContext.Core.Core().V1().Endpoints())

	c.Populator = func(obj runtime.Object, stack *riov1.Stack, os *objectset.ObjectSet) error {
		return populate.ServiceForRouteSet(obj.(*riov1.Router), stack, os)
	}

	return nil
}
