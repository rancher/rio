package servicelabels

import (
	"strings"

	"github.com/rancher/rio/pkg/deploy/stack/input"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
)

func SelectorLabels(stack *input.Stack, service *v1beta1.Service) map[string]string {
	m := ServiceLabels(stack, service)
	return map[string]string{
		"app":     m["app"],
		"version": m["version"],
	}
}

func ServiceLabels(stack *input.Stack, service *v1beta1.Service) map[string]string {
	m := RioOnlyServiceLabels(stack, service)
	m = SafeMerge(m, service.Spec.Labels)
	m["app"] = m["rio.cattle.io/service"]
	m["version"] = m["rio.cattle.io/version"]
	return m
}

func RioOnlyServiceLabels(stack *input.Stack, service *v1beta1.Service) map[string]string {
	labels := map[string]string{
		"rio.cattle.io/service": service.Spec.Revision.ServiceName,
		"rio.cattle.io/version": service.Spec.Revision.Version,
	}
	if stack != nil {
		labels["rio.cattle.io/stack"] = stack.Stack.Name
		labels["rio.cattle.io/workspace"] = stack.Stack.Namespace
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
