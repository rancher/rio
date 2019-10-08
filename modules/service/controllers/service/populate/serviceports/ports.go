package serviceports

import (
	"fmt"
	"strings"

	"github.com/rancher/rio/pkg/services"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Protocol(proto riov1.Protocol) (protocol v1.Protocol) {
	switch proto {
	case riov1.ProtocolUDP:
		protocol = v1.ProtocolUDP
	case riov1.ProtocolSCTP:
		protocol = v1.ProtocolSCTP
	default:
		protocol = v1.ProtocolTCP
	}

	return
}

func NormalizeContainerPort(port riov1.ContainerPort) riov1.ContainerPort {
	if port.Port == 0 {
		port.Port = port.TargetPort
	}

	if port.TargetPort == 0 {
		port.TargetPort = port.Port
	}

	if port.Protocol == "" {
		port.Protocol = riov1.ProtocolHTTP
	}

	return port
}

func ContainerPorts(service *riov1.Service) []riov1.ContainerPort {
	var (
		ports   []riov1.ContainerPort
		portMap = map[string]bool{}
	)

	for _, container := range services.ToNamedContainers(service) {
		for _, port := range container.Ports {
			port = NormalizeContainerPort(port)

			if port.Port == 0 {
				continue
			}

			key := fmt.Sprintf("%v/%v", port.Port, port.Protocol)
			if portMap[key] {
				continue
			}
			portMap[key] = true

			ports = append(ports, port)
		}
	}

	return ports
}

func ServiceNamedPorts(service *riov1.Service) (servicePorts []v1.ServicePort) {
	for _, port := range ContainerPorts(service) {
		servicePort := v1.ServicePort{
			Name:     port.Name,
			Port:     port.Port,
			Protocol: Protocol(port.Protocol),
			TargetPort: intstr.IntOrString{
				IntVal: port.TargetPort,
			},
		}

		if servicePort.Name == "" {
			servicePort.Name = strings.ToLower(fmt.Sprintf("%s-%d", port.Protocol, port.Port))
		}

		servicePorts = append(servicePorts, servicePort)
	}

	return
}
