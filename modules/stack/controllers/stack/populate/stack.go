package populate

import (
	"github.com/rancher/rio/modules/stack/controllers/stack/populate/parse"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
)

func Stack(namespace *v1.Namespace, stack *riov1.Stack, output *objectset.ObjectSet) error {
	if stack.Spec.Template == "" {
		return stackobject.ErrSkipObjectSet
	}

	if stack.Spec.Template != "" {
		if err := parse.Populate(stack, output); err != nil {
			return err
		}
	}

	return nil
}
