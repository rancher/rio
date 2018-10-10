package container

import (
	"strings"

	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
)

func allContainerPorts(con *v1beta1.ContainerConfig) []v1beta1.ExposedPort {
	var eps []v1beta1.ExposedPort
	for _, ep := range con.ExposedPorts {
		ep.Port = 0
		ep.IP = ""
		eps = append(eps, ep)
	}

	for _, portBindings := range con.PortBindings {
		if portBindings.Port > 0 && !strings.EqualFold(portBindings.Protocol, "tcp") {
			portBindings.Port = 0
			portBindings.IP = ""
		}
		eps = append(eps, v1beta1.ExposedPort{
			PortBinding: portBindings,
		})
	}

	return eps
}
