package stackdef

import (
	"fmt"

	"github.com/rancher/rio/pkg/deploy/stackdef/output"
	"github.com/rancher/rio/pkg/deploy/stackdef/populate"
	"github.com/rancher/rio/pkg/deploy/stackdef/populate/ns"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"k8s.io/api/core/v1"
)

func Remove(stack *v1beta1.Stack) error {
	return output.NewDeployment().Deploy(groupID(stack))
}

func Deploy(namespace *v1.Namespace, stack *v1beta1.Stack) error {
	dep := output.NewDeployment()
	if stack.Spec.Template == "" {
		ns.Populate(namespace, stack, dep)
	} else {
		if err := populate.Populate(namespace, stack, dep); err != nil {
			return err
		}
	}

	return dep.Deploy(groupID(stack))
}

func groupID(stack *v1beta1.Stack) string {
	return fmt.Sprintf("stackdef/%s/%s", stack.Namespace, stack.Name)
}
