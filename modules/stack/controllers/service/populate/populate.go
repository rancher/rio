package populate

import (
	"github.com/rancher/rio/modules/stack/controllers/service/populate/autoscale"
	"github.com/rancher/rio/modules/stack/controllers/service/populate/k8sservice"
	"github.com/rancher/rio/modules/stack/controllers/service/populate/podcontrollers"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/serviceset"
	"github.com/rancher/wrangler/pkg/objectset"
)

func Service(configsByName map[string]*v1.Config, volumesByName map[string]*v1.Volume,
	services []*v1.Service, service *v1.Service, os *objectset.ObjectSet) error {
	var err error

	autoscale.Populate(services, os)

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
		if err := podcontrollers.Populate(configsByName, volumesByName, s, os); err != nil {
			return err
		}
	}

	return nil
}
