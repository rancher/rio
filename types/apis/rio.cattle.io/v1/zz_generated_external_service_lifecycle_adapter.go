package v1

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type ExternalServiceLifecycle interface {
	Create(obj *ExternalService) (runtime.Object, error)
	Remove(obj *ExternalService) (runtime.Object, error)
	Updated(obj *ExternalService) (runtime.Object, error)
}

type externalServiceLifecycleAdapter struct {
	lifecycle ExternalServiceLifecycle
}

func (w *externalServiceLifecycleAdapter) HasCreate() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasCreate()
}

func (w *externalServiceLifecycleAdapter) HasFinalize() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasFinalize()
}

func (w *externalServiceLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*ExternalService))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *externalServiceLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*ExternalService))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *externalServiceLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*ExternalService))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewExternalServiceLifecycleAdapter(name string, clusterScoped bool, client ExternalServiceInterface, l ExternalServiceLifecycle) ExternalServiceHandlerFunc {
	adapter := &externalServiceLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *ExternalService) (runtime.Object, error) {
		newObj, err := syncFn(key, obj)
		if o, ok := newObj.(runtime.Object); ok {
			return o, err
		}
		return nil, err
	}
}
