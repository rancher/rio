package services

import (
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
)

func AppAndVersion(service *v1.Service) (string, string) {
	app := service.Spec.App
	version := service.Spec.Version

	if app == "" {
		app = service.Name
	}
	if version == "" {
		version = constants.DefaultServiceVersion
	}

	return app, version
}
