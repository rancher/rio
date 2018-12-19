package v1alpha1

import (
	"github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type ClusterIssuerLifecycle interface {
	Create(obj *v1alpha1.ClusterIssuer) (runtime.Object, error)
	Remove(obj *v1alpha1.ClusterIssuer) (runtime.Object, error)
	Updated(obj *v1alpha1.ClusterIssuer) (runtime.Object, error)
}

type clusterIssuerLifecycleAdapter struct {
	lifecycle ClusterIssuerLifecycle
}

func (w *clusterIssuerLifecycleAdapter) HasCreate() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasCreate()
}

func (w *clusterIssuerLifecycleAdapter) HasFinalize() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasFinalize()
}

func (w *clusterIssuerLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*v1alpha1.ClusterIssuer))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *clusterIssuerLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*v1alpha1.ClusterIssuer))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *clusterIssuerLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*v1alpha1.ClusterIssuer))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewClusterIssuerLifecycleAdapter(name string, clusterScoped bool, client ClusterIssuerInterface, l ClusterIssuerLifecycle) ClusterIssuerHandlerFunc {
	adapter := &clusterIssuerLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *v1alpha1.ClusterIssuer) (runtime.Object, error) {
		newObj, err := syncFn(key, obj)
		if o, ok := newObj.(runtime.Object); ok {
			return o, err
		}
		return nil, err
	}
}
