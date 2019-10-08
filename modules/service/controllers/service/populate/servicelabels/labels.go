package servicelabels

import (
	"strings"

	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/services"
)

var (
	statsPatternAnnotationKey = "sidecar.istio.io/statsInclusionPrefixes"

	defaultEnvoyStatsMatcherInclusionPatterns = []string{
		"http",
		"cluster_manager",
		"listener_manager",
		"http_mixer_filter",
		"tcp_mixer_filter",
		"server",
		"cluster.xds-grpc",
	}
)

func SelectorLabels(service *v1.Service) map[string]string {
	app, version := services.AppAndVersion(service)
	return map[string]string{
		"app":     app,
		"version": version,
	}
}

func ServiceLabels(service *v1.Service) map[string]string {
	return Merge(service.Labels, labels(service), SelectorLabels(service))
}

func ServiceAnnotations(service *v1.Service) map[string]string {
	// user annotations will override ours
	return Merge(annotations(service), service.Annotations)
}

func labels(service *v1.Service) map[string]string {
	return map[string]string{
		"rio.cattle.io/service": service.Name,
	}
}

func annotations(service *v1.Service) map[string]string {
	result := map[string]string{}
	if service.Spec.ServiceMesh != nil && !*service.Spec.ServiceMesh {
		result["rio.cattle.io/mesh"] = "false"
	} else {
		result["rio.cattle.io/mesh"] = "true"
		result[statsPatternAnnotationKey] = strings.Join(defaultEnvoyStatsMatcherInclusionPatterns, ",")
	}
	return result
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
		result = Merge(Merge(base, overlay[0]), overlay[1:]...)
	}

	return result
}
