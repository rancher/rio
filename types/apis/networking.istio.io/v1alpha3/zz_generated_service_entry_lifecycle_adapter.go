package v1alpha3

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type ServiceEntryLifecycle interface {
	Create(obj *ServiceEntry) (runtime.Object, error)
	Remove(obj *ServiceEntry) (runtime.Object, error)
	Updated(obj *ServiceEntry) (runtime.Object, error)
}

type serviceEntryLifecycleAdapter struct {
	lifecycle ServiceEntryLifecycle
}

func (w *serviceEntryLifecycleAdapter) HasCreate() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasCreate()
}

func (w *serviceEntryLifecycleAdapter) HasFinalize() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasFinalize()
}

func (w *serviceEntryLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*ServiceEntry))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *serviceEntryLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*ServiceEntry))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *serviceEntryLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*ServiceEntry))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewServiceEntryLifecycleAdapter(name string, clusterScoped bool, client ServiceEntryInterface, l ServiceEntryLifecycle) ServiceEntryHandlerFunc {
	adapter := &serviceEntryLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *ServiceEntry) (runtime.Object, error) {
		newObj, err := syncFn(key, obj)
		if o, ok := newObj.(runtime.Object); ok {
			return o, err
		}
		return nil, err
	}
}
