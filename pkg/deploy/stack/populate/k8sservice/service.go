package k8sservice

import (
	"github.com/rancher/rio/pkg/deploy/stack/input"
	"github.com/rancher/rio/pkg/deploy/stack/output"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

func Populate(stack *input.Stack, service *riov1.Service, output *output.Deployment) {
	nodePorts(stack, service, output)
	serviceSelector(stack, service, output)
}
