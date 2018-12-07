package v1

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type PublicDomainLifecycle interface {
	Create(obj *PublicDomain) (runtime.Object, error)
	Remove(obj *PublicDomain) (runtime.Object, error)
	Updated(obj *PublicDomain) (runtime.Object, error)
}

type publicDomainLifecycleAdapter struct {
	lifecycle PublicDomainLifecycle
}

func (w *publicDomainLifecycleAdapter) HasCreate() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasCreate()
}

func (w *publicDomainLifecycleAdapter) HasFinalize() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasFinalize()
}

func (w *publicDomainLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*PublicDomain))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *publicDomainLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*PublicDomain))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *publicDomainLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*PublicDomain))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewPublicDomainLifecycleAdapter(name string, clusterScoped bool, client PublicDomainInterface, l PublicDomainLifecycle) PublicDomainHandlerFunc {
	adapter := &publicDomainLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *PublicDomain) (runtime.Object, error) {
		newObj, err := syncFn(key, obj)
		if o, ok := newObj.(runtime.Object); ok {
			return o, err
		}
		return nil, err
	}
}
