package stack

import (
	"context"

	"github.com/rancher/rio/pkg/stack"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/objectset"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "stack-riofile", rContext.Rio.Rio().V1().Stack())
	c.Apply = c.Apply.WithCacheTypes(
		rContext.Rio.Rio().V1().Service(),
		rContext.Core.Core().V1().ConfigMap())

	p := stackPopulator{
		apply: c.Apply,
	}

	c.Populator = p.populate
	return nil
}

type stackPopulator struct {
	apply apply.Apply
}

func (s stackPopulator) populate(obj runtime.Object, ns *corev1.Namespace, os *objectset.ObjectSet) error {
	st := obj.(*riov1.Stack)

	deployStack := stack.NewStack([]byte(st.Spec.Template), st.Spec.Answers)

	if err := deployStack.SetServiceImages(st.Spec.Images); err != nil {
		return err
	}

	objs, err := deployStack.GetObjects()
	if err != nil {
		return err
	}
	os.Add(objs...)
	return nil
}
