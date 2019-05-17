package k8sservice

import (
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/objectset"
)

func Populate(service *riov1.Service, systemNamespace string, os *objectset.ObjectSet) {
	serviceSelector(service, systemNamespace, os)
}
