package populate

import (
	"github.com/rancher/rio/pkg/deploy/stackdef/output"
	"github.com/rancher/rio/pkg/deploy/stackdef/populate/ns"
	"github.com/rancher/rio/pkg/deploy/stackdef/populate/parse"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"k8s.io/api/core/v1"
)

func Populate(namespace *v1.Namespace, stack *v1beta1.Stack, output *output.Deployment) error {
	ns.Populate(namespace, stack, output)
	return parse.Populate(stack, output)
}
