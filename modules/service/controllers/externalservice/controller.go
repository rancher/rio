package externalservice

import (
	"context"

	"github.com/rancher/rio/modules/service/controllers/externalservice/populate"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v12 "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "stack-external-service", rContext.Rio.Rio().V1().ExternalService())
	c.Apply = c.Apply.WithCacheTypes(rContext.Core.Core().V1().Service(),
		rContext.Core.Core().V1().Endpoints())

	p := populator{
		serviceCache: rContext.Rio.Rio().V1().Service().Cache(),
	}

	c.Populator = p.populate
	return nil
}

type populator struct {
	serviceCache v12.ServiceCache
}

func (p populator) populate(obj runtime.Object, namespace *corev1.Namespace, os *objectset.ObjectSet) error {
	return populate.ServiceForExternalService(obj.(*riov1.ExternalService), namespace, os)
}
