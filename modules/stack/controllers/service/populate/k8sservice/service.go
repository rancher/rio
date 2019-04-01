package k8sservice

import (
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/objectset"
)

func Populate(service *riov1.Service, os *objectset.ObjectSet) {
	nodePorts(service, os)
	serviceSelector(service, os)
}
