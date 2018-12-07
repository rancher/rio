package v1

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type RouteSetLifecycle interface {
	Create(obj *RouteSet) (runtime.Object, error)
	Remove(obj *RouteSet) (runtime.Object, error)
	Updated(obj *RouteSet) (runtime.Object, error)
}

type routeSetLifecycleAdapter struct {
	lifecycle RouteSetLifecycle
}

func (w *routeSetLifecycleAdapter) HasCreate() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasCreate()
}

func (w *routeSetLifecycleAdapter) HasFinalize() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasFinalize()
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
	return func(key string, obj *RouteSet) (runtime.Object, error) {
		newObj, err := syncFn(key, obj)
		if o, ok := newObj.(runtime.Object); ok {
			return o, err
		}
		return nil, err
	}
}
