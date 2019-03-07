package v1

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type GitWebHookReceiverLifecycle interface {
	Create(obj *GitWebHookReceiver) (runtime.Object, error)
	Remove(obj *GitWebHookReceiver) (runtime.Object, error)
	Updated(obj *GitWebHookReceiver) (runtime.Object, error)
}

type gitWebHookReceiverLifecycleAdapter struct {
	lifecycle GitWebHookReceiverLifecycle
}

func (w *gitWebHookReceiverLifecycleAdapter) HasCreate() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasCreate()
}

func (w *gitWebHookReceiverLifecycleAdapter) HasFinalize() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasFinalize()
}

func (w *gitWebHookReceiverLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*GitWebHookReceiver))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *gitWebHookReceiverLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*GitWebHookReceiver))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *gitWebHookReceiverLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*GitWebHookReceiver))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewGitWebHookReceiverLifecycleAdapter(name string, clusterScoped bool, client GitWebHookReceiverInterface, l GitWebHookReceiverLifecycle) GitWebHookReceiverHandlerFunc {
	adapter := &gitWebHookReceiverLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *GitWebHookReceiver) (runtime.Object, error) {
		newObj, err := syncFn(key, obj)
		if o, ok := newObj.(runtime.Object); ok {
			return o, err
		}
		return nil, err
	}
}
