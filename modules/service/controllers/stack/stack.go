package stack

import (
	"context"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	controllerName := "stack-riofile"
	c := stackobject.NewGeneratingController(ctx, rContext, controllerName, rContext.Rio.Rio().V1().Stack())
	c.Apply = rContext.Apply.WithSetID(controllerName).WithCacheTypes(
		rContext.Rio.Rio().V1().Service(),
		rContext.Core.Core().V1().ConfigMap())

	p := stackPopulator{}

	c.Populator = p.populate
	return nil
}

type stackPopulator struct{}

func (s stackPopulator) populate(obj runtime.Object, ns *corev1.Namespace, os *objectset.ObjectSet) error {
	st := obj.(*riov1.Stack)
	if st.Spec.Template == "" {
		return nil
	}

	deployStack := stack.NewStack([]byte(st.Spec.Template), st.Spec.Answers)

	if err := deployStack.SetServiceImages(st.Spec.Images); err != nil {
		return err
	}

	objs, err := deployStack.GetObjects()
	if err != nil {
		return err
	}
	accessor := meta.NewAccessor()
	for _, obj := range objs {
		if err := accessor.SetNamespace(obj, st.Namespace); err != nil {
			return err
		}
	}
	os.Add(objs...)
	return nil
}
