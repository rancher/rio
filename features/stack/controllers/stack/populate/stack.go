package populate

import (
	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/features/stack/controllers/stack/populate/ns"
	"github.com/rancher/rio/features/stack/controllers/stack/populate/parse"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"k8s.io/api/core/v1"
)

func Stack(namespace *v1.Namespace, stack *riov1.Stack, output *objectset.ObjectSet) error {
	ns.Populate(namespace, stack, output)
	if stack.Spec.Template != "" {
		if err := parse.Populate(stack, output); err != nil {
			return err
		}
	}

	return nil
}
