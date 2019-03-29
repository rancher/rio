package config

import (
	"context"

	"github.com/rancher/rio/modules/stack/controllers/config/populate"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "stack-config", rContext.Rio.Rio().V1().Config())
	c.Apply.WithCacheTypes(rContext.Core.Core().V1().ConfigMap())

	c.Populator = func(obj runtime.Object, stack *riov1.Stack, os *objectset.ObjectSet) error {
		return populate.Config(stack, obj.(*riov1.Config), os)
	}

	return nil
}
