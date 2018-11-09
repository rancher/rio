package v1alpha3

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type GatewayLifecycle interface {
	Create(obj *Gateway) (runtime.Object, error)
	Remove(obj *Gateway) (runtime.Object, error)
	Updated(obj *Gateway) (runtime.Object, error)
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
	return func(key string, obj *Gateway) (runtime.Object, error) {
		newObj, err := syncFn(key, obj)
		if o, ok := newObj.(runtime.Object); ok {
			return o, err
		}
		return nil, err
	}
}
