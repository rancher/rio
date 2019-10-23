package domains

import (
	"github.com/rancher/rio/pkg/services"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

func IsPublic(svc *riov1.Service) bool {
	for _, container := range services.ToNamedContainers(svc) {
		for _, port := range container.Ports {
			if port.IsExposed() {
				return true
			}
		}
	}
	return false
}

func IsPublicRouter(router *riov1.Router) bool {
	return !router.Spec.Internal
}
