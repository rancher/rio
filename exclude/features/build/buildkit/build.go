package buildkit

import (
	"context"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/controllers/data"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
)

const (
	buildController = "build-controller"
	build           = "build"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		systemNamespace: rContext.SystemNamespace,
		apply: rContext.Apply.WithSetID("build-system-stacks").
			WithStrictCaching().
			WithCacheTypes(rContext.Rio.Rio().V1().Stack()),
	}
	rContext.Core.Core().V1().Namespace().OnChange(ctx, "project-build", h.sync)
	return nil
}

type handler struct {
	systemNamespace string
	apply           apply.Apply
}

func (h handler) sync(key string, obj *v1.Namespace) (*v1.Namespace, error) {
	if obj == nil {
		return nil, nil
	}

	if obj.Labels["rio.cattle.io/project"] != "true" {
		return obj, nil
	}

	//deploy build-controller in rio-system
	if obj.Name == h.systemNamespace {
		stack := data.Stack(buildController, obj.Name, riov1.StackSpec{
			EnableKubernetesResources: true,
		})
		return obj, h.apply.WithOwner(obj).Apply(objectset.NewObjectSet().Add(stack))
	}

	// deploy buildkit, registry and webhook server in each project
	stack := data.Stack(build, obj.Name, riov1.StackSpec{
		EnableKubernetesResources: true,
	})
	return obj, h.apply.WithOwner(obj).Apply(objectset.NewObjectSet().Add(stack))
}
