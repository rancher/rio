package externalservice

import (
	"context"

	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/features/routing/controllers/externalservice/populate"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "routing-external-service", rContext.Rio.ExternalService)
	c.Processor.Client(rContext.Networking.ServiceEntry)

	c.Populator = func(obj runtime.Object, stack *riov1.Stack, os *objectset.ObjectSet) error {
		return populate.ServiceEntry(stack, obj.(*riov1.ExternalService), os)
	}

	return nil
}
