package ports

import (
	"sort"

	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/openapi"
	"github.com/docker/cli/opts"
)

type PortType string

var (
	PortTypeNodePort           = PortType("nodePort")
	PortTypeHostPort           = PortType("hostPort")
	PortTypeExposedPort        = PortType("exposedPort")
	PortTypeVirtualServicePort = PortType("virtualPort")
)

type PortDescription struct {
	Name string
	Protocol string
	PortType PortType

	ServicePort int64
	ContainerPort int64
	VirtualServicePort int64
}

func ReadPorts(service *v1beta1.ServiceUnversionedSpec) []PortDescription {
	names := map[string]bool{}

	eps := service.ExposedPorts
	for _, k := range sortKeys(service.Sidekicks) {
		eps = append(eps, service.Sidekicks[k].ExposedPorts...)
	}

	for i, ep := range eps {
		pd := PortDescription{
			ContainerPort:ep.TargetPort,
			ServicePort:ep.Port,
			Protocol: ep.Protocol,
		}

		if pd.ServicePort == 0 {
			pd.ServicePort =  pd.ContainerPort
		}

		if ep.Port == 0 {
			pd.ServicePort = ep.Port
			pd.PortType = PortTypeExposedPort
		} else if pd.Protocol == "http" {
			pd.ServicePort = ep.Port
			pd.PortType = PortTypeExposedPort

		}
		if pd.ServicePort == 0 {
			pd.ServicePort = pd.ContainerPort
		} else if pd.Protocol == "http" {
			pd.PortType = PortTypeVirtualServicePort
			pd.VirtualServicePort = ep.Port

		}

		if pd.Protocol == "http" {
			pd.PortType =
		}
		if ep.Name
	}
}

func sortKeys(m map[string]v1beta1.SidekickConfig) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
