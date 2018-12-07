package v1

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type StackLifecycle interface {
	Create(obj *Stack) (runtime.Object, error)
	Remove(obj *Stack) (runtime.Object, error)
	Updated(obj *Stack) (runtime.Object, error)
}

type stackLifecycleAdapter struct {
	lifecycle StackLifecycle
}

func (w *stackLifecycleAdapter) HasCreate() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasCreate()
}

func (w *stackLifecycleAdapter) HasFinalize() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasFinalize()
}

func (w *stackLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*Stack))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *stackLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*Stack))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *stackLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*Stack))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewStackLifecycleAdapter(name string, clusterScoped bool, client StackInterface, l StackLifecycle) StackHandlerFunc {
	adapter := &stackLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *Stack) (runtime.Object, error) {
		newObj, err := syncFn(key, obj)
		if o, ok := newObj.(runtime.Object); ok {
			return o, err
		}
		return nil, err
	}
}
