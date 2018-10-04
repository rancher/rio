package v3

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type SourceCodeCredentialLifecycle interface {
	Create(obj *SourceCodeCredential) (*SourceCodeCredential, error)
	Remove(obj *SourceCodeCredential) (*SourceCodeCredential, error)
	Updated(obj *SourceCodeCredential) (*SourceCodeCredential, error)
}

type sourceCodeCredentialLifecycleAdapter struct {
	lifecycle SourceCodeCredentialLifecycle
}

func (w *sourceCodeCredentialLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*SourceCodeCredential))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *sourceCodeCredentialLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*SourceCodeCredential))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *sourceCodeCredentialLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*SourceCodeCredential))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewSourceCodeCredentialLifecycleAdapter(name string, clusterScoped bool, client SourceCodeCredentialInterface, l SourceCodeCredentialLifecycle) SourceCodeCredentialHandlerFunc {
	adapter := &sourceCodeCredentialLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *SourceCodeCredential) error {
		if obj == nil {
			return syncFn(key, nil)
		}
		return syncFn(key, obj)
	}
}
