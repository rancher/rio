package services

import (
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

func AppAndVersion(service *v1.Service) (string, string) {
	app := service.Spec.App
	version := service.Spec.Version

	if app == "" {
		app = service.Name
	}
	if version == "" {
		if len(service.UID) < 8 {
			version = string(service.UID)
		} else {
			version = string(service.UID)[:8]
		}
	}

	return app, version
}

func RootContainerName(service *v1.Service) string {
	appName, _ := AppAndVersion(service)
	return appName
}

func containerIsValid(container *v1.Container) bool {
	return container.Image != "" || container.ImageBuild != nil
}

func ToNamedContainers(service *v1.Service) (result []v1.NamedContainer) {
	if containerIsValid(&service.Spec.Container) {
		result = append(result, v1.NamedContainer{
			Name:      RootContainerName(service),
			Container: service.Spec.Container,
		})
	}

	result = append(result, service.Spec.Sidecars...)
	return
}
