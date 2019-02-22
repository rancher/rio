package v1

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type GitWebHookExecutionLifecycle interface {
	Create(obj *GitWebHookExecution) (runtime.Object, error)
	Remove(obj *GitWebHookExecution) (runtime.Object, error)
	Updated(obj *GitWebHookExecution) (runtime.Object, error)
}

type gitWebHookExecutionLifecycleAdapter struct {
	lifecycle GitWebHookExecutionLifecycle
}

func (w *gitWebHookExecutionLifecycleAdapter) HasCreate() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasCreate()
}

func (w *gitWebHookExecutionLifecycleAdapter) HasFinalize() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasFinalize()
}

func (w *gitWebHookExecutionLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*GitWebHookExecution))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *gitWebHookExecutionLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*GitWebHookExecution))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *gitWebHookExecutionLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*GitWebHookExecution))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewGitWebHookExecutionLifecycleAdapter(name string, clusterScoped bool, client GitWebHookExecutionInterface, l GitWebHookExecutionLifecycle) GitWebHookExecutionHandlerFunc {
	adapter := &gitWebHookExecutionLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *GitWebHookExecution) (runtime.Object, error) {
		newObj, err := syncFn(key, obj)
		if o, ok := newObj.(runtime.Object); ok {
			return o, err
		}
		return nil, err
	}
}
