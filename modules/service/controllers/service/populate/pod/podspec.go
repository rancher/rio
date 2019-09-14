package pod

import (
	"github.com/rancher/rio/modules/service/controllers/service/populate/rbac"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	f             = false
	t             = true
	defaultCPU    = resource.MustParse("50m")
	defaultMemory = resource.MustParse("64Mi")
)

func podSpec(service *riov1.Service, systemNamespace string) v1.PodSpec {
	podSpec := v1.PodSpec{
		DNSConfig:          podDNS(service),
		DNSPolicy:          service.Spec.DNSPolicy,
		HostAliases:        service.Spec.HostAliases,
		Hostname:           service.Spec.Hostname,
		HostNetwork:        service.Spec.HostNetwork,
		ServiceAccountName: rbac.ServiceAccountName(service),
		EnableServiceLinks: &f,
		Containers:         containers(service, systemNamespace, false),
		InitContainers:     containers(service, systemNamespace, true),
		Volumes:            volumes(service),
		Affinity:           service.Spec.Affinity,
		ImagePullSecrets:   service.Spec.ImagePullSecrets,
	}

	if podSpec.ServiceAccountName == "" {
		podSpec.AutomountServiceAccountToken = &f
	} else {
		podSpec.AutomountServiceAccountToken = &t
	}

	return podSpec
}

func podDNS(service *riov1.Service) *v1.PodDNSConfig {
	if len(service.Spec.PodDNSConfig.Options) == 0 &&
		len(service.Spec.PodDNSConfig.Nameservers) == 0 &&
		len(service.Spec.PodDNSConfig.Searches) == 0 {
		return nil
	}

	var options []v1.PodDNSConfigOption
	for _, opt := range service.Spec.PodDNSConfig.Options {
		options = append(options, v1.PodDNSConfigOption{
			Name:  opt.Name,
			Value: opt.Value,
		})
	}
	return &v1.PodDNSConfig{
		Options:     options,
		Nameservers: service.Spec.PodDNSConfig.Nameservers,
		Searches:    service.Spec.PodDNSConfig.Searches,
	}
}
