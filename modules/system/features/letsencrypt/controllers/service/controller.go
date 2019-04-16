package service

import (
	"context"

	"github.com/rancher/rio/modules/system/features/letsencrypt/controllers/service/populate"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "letsencrypt-service", rContext.Rio.Rio().V1().Service())
	c.Apply = c.Apply.WithStrictCaching().WithCacheTypes(rContext.Networking.Networking().V1alpha3().DestinationRule())

	c.Populator = func(obj runtime.Object, namespace *corev1.Namespace, os *objectset.ObjectSet) error {
		return populate.DestinationRules(obj.(*riov1.Service), namespace.Name, os)
	}

	return nil
}
