package v1

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type ConfigLifecycle interface {
	Create(obj *Config) (runtime.Object, error)
	Remove(obj *Config) (runtime.Object, error)
	Updated(obj *Config) (runtime.Object, error)
}

type configLifecycleAdapter struct {
	lifecycle ConfigLifecycle
}

func (w *configLifecycleAdapter) HasCreate() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasCreate()
}

func (w *configLifecycleAdapter) HasFinalize() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasFinalize()
}

func (w *configLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*Config))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *configLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*Config))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *configLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*Config))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewConfigLifecycleAdapter(name string, clusterScoped bool, client ConfigInterface, l ConfigLifecycle) ConfigHandlerFunc {
	adapter := &configLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *Config) (runtime.Object, error) {
		newObj, err := syncFn(key, obj)
		if o, ok := newObj.(runtime.Object); ok {
			return o, err
		}
		return nil, err
	}
}
