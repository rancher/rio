package serviceports

import (
	"fmt"
	"strings"

	"github.com/rancher/rio/modules/stack/controllers/service/populate/containerlist"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func AllExposedPorts(con *riov1.ContainerConfig) []riov1.ExposedPort {
	var eps []riov1.ExposedPort
	eps = append(eps, con.ExposedPorts...)

	for _, portBindings := range con.PortBindings {
		pb := riov1.PortBinding{
			TargetPort: portBindings.TargetPort,
			Protocol:   portBindings.Protocol,
		}
		if portBindings.Port != 0 && strings.EqualFold(pb.Protocol, "tcp") {
			pb.Port = portBindings.Port
		}
		eps = append(eps, riov1.ExposedPort{
			PortBinding: pb,
		})
	}

	return eps
}

func ServiceNamedPorts(service *riov1.Service) ([]v1.ServicePort, string) {
	var result []v1.ServicePort

	var eps []riov1.ExposedPort
	for _, con := range containerlist.ForService(service) {
		eps = append(eps, AllExposedPorts(con)...)
	}

	ip := ""
	portsDefined := map[string]bool{}
	names := map[string]bool{}
	for _, port := range eps {
		if port.Port == 0 {
			port.Port = port.TargetPort
		}

		if port.IP != "" {
			ip = port.IP
		}

		name := ""
		defName := fmt.Sprintf("%s-%d-%d", port.Protocol, port.Port, port.TargetPort)
		if port.Name == "" {
			name = defName
		} else {
			name = port.Name
		}
		if names[name] || portsDefined[defName] {
			continue
		}

		portsDefined[defName] = true
		names[name] = true

		servicePort := v1.ServicePort{
			Name:       name,
			TargetPort: intstr.FromInt(int(port.TargetPort)),
			Port:       int32(port.Port),
			Protocol:   v1.ProtocolTCP,
		}

		if strings.EqualFold(port.Protocol, "udp") {
			servicePort.Protocol = v1.ProtocolUDP
		}

		result = append(result, servicePort)
	}

	return result, ip
}
