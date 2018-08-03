package v1beta1

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type VolumeLifecycle interface {
	Create(obj *Volume) (*Volume, error)
	Remove(obj *Volume) (*Volume, error)
	Updated(obj *Volume) (*Volume, error)
}

type volumeLifecycleAdapter struct {
	lifecycle VolumeLifecycle
}

func (w *volumeLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*Volume))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *volumeLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*Volume))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *volumeLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*Volume))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewVolumeLifecycleAdapter(name string, clusterScoped bool, client VolumeInterface, l VolumeLifecycle) VolumeHandlerFunc {
	adapter := &volumeLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *Volume) error {
		if obj == nil {
			return syncFn(key, nil)
		}
		return syncFn(key, obj)
	}
}
