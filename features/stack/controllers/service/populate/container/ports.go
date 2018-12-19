package container

import (
	"strings"

	"github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

func allContainerPorts(con *v1.ContainerConfig) []v1.ExposedPort {
	var eps []v1.ExposedPort
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
		eps = append(eps, v1.ExposedPort{
			PortBinding: portBindings,
		})
	}

	return eps
}
