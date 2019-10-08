package gitcommit

import (
	webhookv1 "github.com/rancher/gitwatcher/pkg/apis/gitwatcher.cattle.io/v1"
	"github.com/rancher/rio/modules/build/pkg"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (h Handler) onChangeStack(key string, obj *webhookv1.GitCommit, gitWatcher *webhookv1.GitWatcher) (*webhookv1.GitCommit, error) {
	stack, err := h.stacks.Cache().Get(obj.Namespace, gitWatcher.Annotations[pkg.StackLabel])
	if err != nil {
		if errors.IsNotFound(err) {
			return obj, nil
		}
		return obj, err
	}

	// if git commit is from different branch do no-op
	if obj.Spec.Branch != "" && obj.Spec.Branch != stack.Spec.Build.Branch {
		return obj, nil
	}

	if obj.Spec.Commit == "" {
		return obj, nil
	}

	if stack.Status.Revision != obj.Spec.Commit {
		stack.Status.Revision = obj.Spec.Commit
		if _, err := h.stacks.Update(stack); err != nil {
			return nil, err
		}
	}

	return obj, nil
}
