package v1alpha3

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type VirtualServiceLifecycle interface {
	Create(obj *VirtualService) (*VirtualService, error)
	Remove(obj *VirtualService) (*VirtualService, error)
	Updated(obj *VirtualService) (*VirtualService, error)
}

type virtualServiceLifecycleAdapter struct {
	lifecycle VirtualServiceLifecycle
}

func (w *virtualServiceLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*VirtualService))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *virtualServiceLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*VirtualService))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *virtualServiceLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*VirtualService))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewVirtualServiceLifecycleAdapter(name string, clusterScoped bool, client VirtualServiceInterface, l VirtualServiceLifecycle) VirtualServiceHandlerFunc {
	adapter := &virtualServiceLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *VirtualService) error {
		if obj == nil {
			return syncFn(key, nil)
		}
		return syncFn(key, obj)
	}
}
