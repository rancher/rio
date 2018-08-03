package v1beta1

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type ConfigLifecycle interface {
	Create(obj *Config) (*Config, error)
	Remove(obj *Config) (*Config, error)
	Updated(obj *Config) (*Config, error)
}

type configLifecycleAdapter struct {
	lifecycle ConfigLifecycle
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
	return func(key string, obj *Config) error {
		if obj == nil {
			return syncFn(key, nil)
		}
		return syncFn(key, obj)
	}
}
