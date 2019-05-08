package servicelabels

import (
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/services"
)

func SelectorLabels(service *v1.Service) map[string]string {
	return labels(service)
}

func ServiceLabels(service *v1.Service) map[string]string {
	return Merge(service.Labels, labels(service))
}

func labels(service *v1.Service) map[string]string {
	app, version := services.AppAndVersion(service)
	labels := map[string]string{
		"app":     app,
		"version": version,
	}

	return labels
}

func Merge(base map[string]string, overlay ...map[string]string) map[string]string {
	result := map[string]string{}
	for k, v := range base {
		result[k] = v
	}

	i := len(overlay)
	switch {
	case i == 1:
		for k, v := range overlay[0] {
			result[k] = v
		}
	case i > 1:
		result = Merge(Merge(base, overlay[1]), overlay[2:]...)
	}

	return result
}
