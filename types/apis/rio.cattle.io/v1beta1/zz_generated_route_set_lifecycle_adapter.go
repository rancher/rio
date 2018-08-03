package v1beta1

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type RouteSetLifecycle interface {
	Create(obj *RouteSet) (*RouteSet, error)
	Remove(obj *RouteSet) (*RouteSet, error)
	Updated(obj *RouteSet) (*RouteSet, error)
}

type routeSetLifecycleAdapter struct {
	lifecycle RouteSetLifecycle
}

func (w *routeSetLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*RouteSet))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *routeSetLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*RouteSet))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *routeSetLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*RouteSet))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewRouteSetLifecycleAdapter(name string, clusterScoped bool, client RouteSetInterface, l RouteSetLifecycle) RouteSetHandlerFunc {
	adapter := &routeSetLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *RouteSet) error {
		if obj == nil {
			return syncFn(key, nil)
		}
		return syncFn(key, obj)
	}
}
