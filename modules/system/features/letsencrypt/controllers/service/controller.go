package service

import (
	"context"

	"github.com/rancher/rio/modules/system/features/letsencrypt/controllers/service/populate"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "letsencrypt-service", rContext.Rio.Rio().V1().Service())
	c.Apply.WithCacheTypes(rContext.Networking.Networking().V1alpha3().DestinationRule())

	c.Populator = func(obj runtime.Object, stack *riov1.Stack, os *objectset.ObjectSet) error {
		return populate.DestinationRules(obj.(*riov1.Service), os)
	}

	return nil
}
