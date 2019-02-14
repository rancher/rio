package populate

import (
	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/features/stack/controllers/service/populate/autoscale"
	"github.com/rancher/rio/features/stack/controllers/service/populate/k8sservice"
	"github.com/rancher/rio/features/stack/controllers/service/populate/podcontrollers"
	"github.com/rancher/rio/pkg/serviceset"
	v1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

func Service(stack *v1.Stack, configsByName map[string]*v1.Config, volumesByName map[string]*v1.Volume,
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
		k8sservice.Populate(stack, s, os)
		if err := podcontrollers.Populate(stack, configsByName, volumesByName, s, os); err != nil {
			return err
		}
	}

	return nil
}
