package stack

import (
	"context"
	"fmt"
	"sync"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/kv"
	"github.com/rancher/wrangler/pkg/objectset"
	"github.com/rancher/wrangler/pkg/relatedresource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	controllerName := "stack-riofile"
	c := stackobject.NewGeneratingController(ctx, rContext, controllerName, rContext.Rio.Rio().V1().Stack())
	c.Apply = rContext.Apply.WithSetID(controllerName).WithCacheTypes(
		rContext.Rio.Rio().V1().Service(),
		rContext.Rio.Rio().V1().ExternalService(),
		rContext.Core.Core().V1().ConfigMap(),
		rContext.Rio.Rio().V1().Router())

	p := stackPopulator{}

	c.Populator = p.populate

	h := handler{
		stackMap: &sync.Map{},
	}
	relatedresource.Watch(ctx, "stack-reenqueue-changes", h.resolve,
		rContext.Rio.Rio().V1().Stack(),
		rContext.Rio.Rio().V1().Service(),
		rContext.Rio.Rio().V1().ExternalService(),
		rContext.Rio.Rio().V1().Router(),
		rContext.Core.Core().V1().ConfigMap())

	return nil
}

type handler struct {
	stackMap *sync.Map
}

func (h handler) resolve(namespace, name string, obj runtime.Object) ([]relatedresource.Key, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)
	if obj != nil {
		n, ns := ownedByStack(obj)
		if n == "" || ns == "" {
			return nil, nil
		}
		h.stackMap.Store(key, fmt.Sprintf("%s/%s", ns, n))
	}

	value, ok := h.stackMap.Load(key)
	if ok {
		stackNs, stackName := kv.Split(value.(string), "/")
		if obj == nil {
			h.stackMap.Delete(key)
		}

		if stackNs != "" && stackName != "" {
			return []relatedresource.Key{
				{
					Namespace: stackNs,
					Name:      stackName,
				},
			}, nil
		}
	}

	return nil, nil
}

func ownedByStack(obj runtime.Object) (string, string) {
	meta, err := meta.Accessor(obj)
	if err != nil {
		return "", ""
	}

	if meta.GetAnnotations()["objectset.rio.cattle.io/owner-gvk"] == "rio.cattle.io/v1, Kind=Stack" {
		return meta.GetAnnotations()["objectset.rio.cattle.io/owner-name"], meta.GetAnnotations()["objectset.rio.cattle.io/owner-namespace"]
	}
	return "", ""
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
