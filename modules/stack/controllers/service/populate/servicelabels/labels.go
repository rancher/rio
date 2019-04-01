package servicelabels

import (
	"strings"

	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

func SelectorLabels(service *v1.Service) map[string]string {
	m := ServiceLabels(service)
	return map[string]string{
		"app":     m["app"],
		"version": m["version"],
	}
}

func ServiceLabels(service *v1.Service) map[string]string {
	m := RioOnlyServiceLabels(service)
	m = SafeMerge(m, service.Spec.Labels)
	m["app"] = service.Name
	m["version"] = m["rio.cattle.io/version"]
	return m
}

func RioOnlyServiceLabels(service *v1.Service) map[string]string {
	labels := map[string]string{
		"rio.cattle.io/service": service.Spec.Revision.ServiceName,
		"rio.cattle.io/version": service.Spec.Revision.Version,
	}
	if service.Spec.Revision.ParentService == "" {
		labels["rio.cattle.io/service-name"] = service.Spec.Revision.ServiceName
	} else {
		labels["rio.cattle.io/service-name"] = service.Name
	}

	return labels
}

func SafeMerge(base, overlay map[string]string) map[string]string {
	result := map[string]string{}
	for k, v := range base {
		result[k] = v
	}

	for k, v := range overlay {
		if strings.HasPrefix(k, "rio.cattle.io") {
			continue
		}
		result[k] = v
	}

	return result
}
