package v1

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type ServiceScaleRecommendationLifecycle interface {
	Create(obj *ServiceScaleRecommendation) (runtime.Object, error)
	Remove(obj *ServiceScaleRecommendation) (runtime.Object, error)
	Updated(obj *ServiceScaleRecommendation) (runtime.Object, error)
}

type serviceScaleRecommendationLifecycleAdapter struct {
	lifecycle ServiceScaleRecommendationLifecycle
}

func (w *serviceScaleRecommendationLifecycleAdapter) HasCreate() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasCreate()
}

func (w *serviceScaleRecommendationLifecycleAdapter) HasFinalize() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasFinalize()
}

func (w *serviceScaleRecommendationLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*ServiceScaleRecommendation))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *serviceScaleRecommendationLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*ServiceScaleRecommendation))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *serviceScaleRecommendationLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*ServiceScaleRecommendation))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewServiceScaleRecommendationLifecycleAdapter(name string, clusterScoped bool, client ServiceScaleRecommendationInterface, l ServiceScaleRecommendationLifecycle) ServiceScaleRecommendationHandlerFunc {
	adapter := &serviceScaleRecommendationLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *ServiceScaleRecommendation) (runtime.Object, error) {
		newObj, err := syncFn(key, obj)
		if o, ok := newObj.(runtime.Object); ok {
			return o, err
		}
		return nil, err
	}
}
