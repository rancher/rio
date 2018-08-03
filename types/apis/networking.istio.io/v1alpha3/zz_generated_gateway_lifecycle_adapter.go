package v1alpha3

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type GatewayLifecycle interface {
	Create(obj *Gateway) (*Gateway, error)
	Remove(obj *Gateway) (*Gateway, error)
	Updated(obj *Gateway) (*Gateway, error)
}

type gatewayLifecycleAdapter struct {
	lifecycle GatewayLifecycle
}

func (w *gatewayLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*Gateway))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *gatewayLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*Gateway))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *gatewayLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*Gateway))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewGatewayLifecycleAdapter(name string, clusterScoped bool, client GatewayInterface, l GatewayLifecycle) GatewayHandlerFunc {
	adapter := &gatewayLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *Gateway) error {
		if obj == nil {
			return syncFn(key, nil)
		}
		return syncFn(key, obj)
	}
}
