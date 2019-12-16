package gitcommit

import (
	webhookv1 "github.com/rancher/gitwatcher/pkg/apis/gitwatcher.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (h Handler) onChangeStack(key string, obj *webhookv1.GitCommit, gitWatcher *webhookv1.GitWatcher) (*webhookv1.GitCommit, error) {
	stack, err := h.stacks.Cache().Get(obj.Namespace, gitWatcher.Annotations[constants.StackLabel])
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

	ref := obj.Spec.Commit
	if ref == "" {
		ref = obj.Spec.Tag
	}
	if ref == "" {
		return obj, nil
	}

	if stack.Status.Revision != ref {
		stack.Status.Revision = ref
		if _, err := h.stacks.UpdateStatus(stack); err != nil {
			return nil, err
		}
	}

	return obj, nil
}
