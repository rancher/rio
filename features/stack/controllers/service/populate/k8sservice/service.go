package k8sservice

import (
	"github.com/rancher/norman/pkg/objectset"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

func Populate(stack *riov1.Stack, service *riov1.Service, os *objectset.ObjectSet) {
	nodePorts(stack, service, os)
	serviceSelector(stack, service, os)
}
