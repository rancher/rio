package containerlist

import (
	"github.com/rancher/rio/pkg/deploy/stack/populate/sidekick"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
)

func ForService(service *v1beta1.Service) []*v1beta1.ContainerConfig {
	var result []*v1beta1.ContainerConfig
	result = append(result, &service.Spec.ContainerConfig)

	for _, k := range sidekick.SortedKeys(service.Spec.Sidekicks) {
		sk := service.Spec.Sidekicks[k]
		result = append(result, &sk.ContainerConfig)
	}

	return result
}
