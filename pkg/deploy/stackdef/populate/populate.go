package populate

import (
	"github.com/rancher/rio/pkg/deploy/stackdef/output"
	"github.com/rancher/rio/pkg/deploy/stackdef/populate/ns"
	"github.com/rancher/rio/pkg/deploy/stackdef/populate/parse"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"k8s.io/api/core/v1"
)

func Populate(namespace *v1.Namespace, stack *riov1.Stack, output *output.Deployment) error {
	ns.Populate(namespace, stack, output)
	return parse.Populate(stack, output)
}
