package serviceports

import (
	"fmt"
	"strings"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "k8s.io/api/core/v1"
)

func ServiceNamedPorts(service *riov1.Service) ([]v1.ServicePort, string) {
	var (
		servicePorts []v1.ServicePort
		ip           string
	)

	for _, port := range service.Spec.Ports {
		if ip == "" {
			ip = port.IP
		}

		if port.Protocol == "" {
			port.Protocol = riov1.ProtocolHTTP
		}

		servicePort := v1.ServicePort{
			Name:       port.Name,
			Port:       port.Port,
			TargetPort: port.TargetPort,
			NodePort:   port.NodePort,
		}

		switch port.Protocol {
		case riov1.ProtocolUDP:
			servicePort.Protocol = v1.ProtocolUDP
		case riov1.ProtocolSCTP:
			servicePort.Protocol = v1.ProtocolSCTP
		default:
			servicePort.Protocol = v1.ProtocolTCP
		}

		if servicePort.Name == "" {
			servicePort.Name = strings.ToLower(fmt.Sprintf("%s-%d", port.Protocol, port.Port))
		}

		servicePorts = append(servicePorts, servicePort)
	}

	return servicePorts, ip
}
