package v1

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type VolumeLifecycle interface {
	Create(obj *Volume) (runtime.Object, error)
	Remove(obj *Volume) (runtime.Object, error)
	Updated(obj *Volume) (runtime.Object, error)
}

type volumeLifecycleAdapter struct {
	lifecycle VolumeLifecycle
}

func (w *volumeLifecycleAdapter) HasCreate() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasCreate()
}

func (w *volumeLifecycleAdapter) HasFinalize() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasFinalize()
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
	return func(key string, obj *Volume) (runtime.Object, error) {
		newObj, err := syncFn(key, obj)
		if o, ok := newObj.(runtime.Object); ok {
			return o, err
		}
		return nil, err
	}
}
