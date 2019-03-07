package v1alpha1

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type BuildLifecycle interface {
	Create(obj *Build) (runtime.Object, error)
	Remove(obj *Build) (runtime.Object, error)
	Updated(obj *Build) (runtime.Object, error)
}

type buildLifecycleAdapter struct {
	lifecycle BuildLifecycle
}

func (w *buildLifecycleAdapter) HasCreate() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasCreate()
}

func (w *buildLifecycleAdapter) HasFinalize() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasFinalize()
}

func (w *buildLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*Build))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *buildLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*Build))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *buildLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*Build))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewBuildLifecycleAdapter(name string, clusterScoped bool, client BuildInterface, l BuildLifecycle) BuildHandlerFunc {
	adapter := &buildLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *Build) (runtime.Object, error) {
		newObj, err := syncFn(key, obj)
		if o, ok := newObj.(runtime.Object); ok {
			return o, err
		}
		return nil, err
	}
}
