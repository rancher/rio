package v1beta1

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type ServiceLifecycle interface {
	Create(obj *Service) (*Service, error)
	Remove(obj *Service) (*Service, error)
	Updated(obj *Service) (*Service, error)
}

type serviceLifecycleAdapter struct {
	lifecycle ServiceLifecycle
}

func (w *serviceLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*Service))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *serviceLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*Service))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *serviceLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*Service))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewServiceLifecycleAdapter(name string, clusterScoped bool, client ServiceInterface, l ServiceLifecycle) ServiceHandlerFunc {
	adapter := &serviceLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *Service) error {
		if obj == nil {
			return syncFn(key, nil)
		}
		return syncFn(key, obj)
	}
}
