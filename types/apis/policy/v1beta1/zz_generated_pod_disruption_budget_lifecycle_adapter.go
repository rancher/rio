package v1beta1

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

type PodDisruptionBudgetLifecycle interface {
	Create(obj *v1beta1.PodDisruptionBudget) (runtime.Object, error)
	Remove(obj *v1beta1.PodDisruptionBudget) (runtime.Object, error)
	Updated(obj *v1beta1.PodDisruptionBudget) (runtime.Object, error)
}

type podDisruptionBudgetLifecycleAdapter struct {
	lifecycle PodDisruptionBudgetLifecycle
}

func (w *podDisruptionBudgetLifecycleAdapter) HasCreate() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasCreate()
}

func (w *podDisruptionBudgetLifecycleAdapter) HasFinalize() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasFinalize()
}

func (w *podDisruptionBudgetLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*v1beta1.PodDisruptionBudget))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *podDisruptionBudgetLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*v1beta1.PodDisruptionBudget))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *podDisruptionBudgetLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*v1beta1.PodDisruptionBudget))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewPodDisruptionBudgetLifecycleAdapter(name string, clusterScoped bool, client PodDisruptionBudgetInterface, l PodDisruptionBudgetLifecycle) PodDisruptionBudgetHandlerFunc {
	adapter := &podDisruptionBudgetLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *v1beta1.PodDisruptionBudget) (runtime.Object, error) {
		newObj, err := syncFn(key, obj)
		if o, ok := newObj.(runtime.Object); ok {
			return o, err
		}
		return nil, err
	}
}
