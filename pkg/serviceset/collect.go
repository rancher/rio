package serviceset

import (
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	services2 "github.com/rancher/rio/pkg/services"
)

func CollectionServices(services []*riov1.Service) (Services, error) {
	result := Services{}
	for _, svc := range services {
		app, _ := services2.AppAndVersion(svc)

		serviceSet, ok := result[app]
		if !ok {
			serviceSet = &ServiceSet{}
			result[app] = serviceSet
		}
		serviceSet.Revisions = append(serviceSet.Revisions, svc)
	}
	return result, nil
}
