package populate

import (
	"github.com/rancher/rio/modules/service/controllers/service/populate/k8sservice"
	"github.com/rancher/rio/modules/service/controllers/service/populate/podcontrollers"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/serviceset"
	"github.com/rancher/wrangler/pkg/objectset"
)

func Service(services []*v1.Service, service *v1.Service, os *objectset.ObjectSet) error {
	var err error

	serviceSets, err := serviceset.CollectionServices(services)
	if err != nil {
		return err
	}

	serviceSet, ok := serviceSets[service.Name]
	if !ok {
		return nil
	}

	for _, s := range serviceSet.List() {
		k8sservice.Populate(s, os)
		if err := podcontrollers.Populate(s, os); err != nil {
			return err
		}
	}

	return nil
}
