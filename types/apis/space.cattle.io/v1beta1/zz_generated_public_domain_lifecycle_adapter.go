package v1beta1

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type PublicDomainLifecycle interface {
	Create(obj *PublicDomain) (*PublicDomain, error)
	Remove(obj *PublicDomain) (*PublicDomain, error)
	Updated(obj *PublicDomain) (*PublicDomain, error)
}

type publicDomainLifecycleAdapter struct {
	lifecycle PublicDomainLifecycle
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
	return func(key string, obj *PublicDomain) error {
		if obj == nil {
			return syncFn(key, nil)
		}
		return syncFn(key, obj)
	}
}
