package populate

import (
	"github.com/rancher/rio/modules/service/controllers/service/populate/k8sservice"
	"github.com/rancher/rio/modules/service/controllers/service/populate/podcontrollers"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/objectset"
)

func Service(service *v1.Service, os *objectset.ObjectSet) error {
	k8sservice.Populate(service, os)
	return podcontrollers.Populate(service, os)
}
