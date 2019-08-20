package gitcommit

import (
	"fmt"

	webhookv1 "github.com/rancher/gitwatcher/pkg/apis/gitwatcher.cattle.io/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (h Handler) onChangeStack(key string, obj *webhookv1.GitCommit, gitWatcher *webhookv1.GitWatcher) (*webhookv1.GitCommit, error) {
	if gitWatcher.Status.FirstCommit == "" {
		gitWatcher, err := h.gitWatcherClient.Get(gitWatcher.Namespace, gitWatcher.Name, v1.GetOptions{})
		if err != nil {
			return nil, err
		}
		if gitWatcher.Status.FirstCommit == "" {
			return obj, fmt.Errorf("waiting for gitWatcher first commit on %s/%s", gitWatcher.Namespace, gitWatcher.Name)
		}
	}

	stack, err := h.stacks.Cache().Get(obj.Namespace, gitWatcher.Annotations["objectset.rio.cattle.io/owner-name"])
	if err != nil {
		if errors.IsNotFound(err) {
			return obj, nil
		}
		return obj, err
	}

	if obj.Spec.Commit == gitWatcher.Status.FirstCommit {
		if stack.Status.Revision == "" && stack.Status.Revision != gitWatcher.Status.FirstCommit {
			stack = stack.DeepCopy()
			stack.Status.Revision = gitWatcher.Status.FirstCommit
			_, err := h.stacks.Update(stack)
			return obj, err
		}
		return obj, nil
	}

	return obj, webhookv1.GitWebHookExecutionConditionHandled.Once(obj, func() (runtime.Object, error) {
		// if git commit is from different branch do no-op
		if obj.Spec.Branch != "" && obj.Spec.Branch != stack.Spec.Build.Branch {
			return obj, nil
		}

		if obj.Spec.Commit == "" {
			return obj, nil
		}

		stack.Status.Revision = obj.Spec.Commit
		if _, err := h.stacks.Update(stack); err != nil {
			return nil, err
		}

		return obj, nil
	})
}
