package v1beta1

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

type CustomResourceDefinitionLifecycle interface {
	Create(obj *v1beta1.CustomResourceDefinition) (runtime.Object, error)
	Remove(obj *v1beta1.CustomResourceDefinition) (runtime.Object, error)
	Updated(obj *v1beta1.CustomResourceDefinition) (runtime.Object, error)
}

type customResourceDefinitionLifecycleAdapter struct {
	lifecycle CustomResourceDefinitionLifecycle
}

func (w *customResourceDefinitionLifecycleAdapter) HasCreate() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasCreate()
}

func (w *customResourceDefinitionLifecycleAdapter) HasFinalize() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasFinalize()
}

func (w *customResourceDefinitionLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*v1beta1.CustomResourceDefinition))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *customResourceDefinitionLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*v1beta1.CustomResourceDefinition))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *customResourceDefinitionLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*v1beta1.CustomResourceDefinition))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewCustomResourceDefinitionLifecycleAdapter(name string, clusterScoped bool, client CustomResourceDefinitionInterface, l CustomResourceDefinitionLifecycle) CustomResourceDefinitionHandlerFunc {
	adapter := &customResourceDefinitionLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *v1beta1.CustomResourceDefinition) (runtime.Object, error) {
		newObj, err := syncFn(key, obj)
		if o, ok := newObj.(runtime.Object); ok {
			return o, err
		}
		return nil, err
	}
}
