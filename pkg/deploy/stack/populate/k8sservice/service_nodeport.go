package k8sservice

import (
	"fmt"

	"github.com/rancher/rio/pkg/deploy/stack/input"
	"github.com/rancher/rio/pkg/deploy/stack/output"
	"github.com/rancher/rio/pkg/deploy/stack/populate/containerlist"
	"github.com/rancher/rio/pkg/deploy/stack/populate/servicelabels"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func nodePorts(stack *input.Stack, service *riov1.Service, output *output.Deployment) {
	labels := servicelabels.SelectorLabels(stack, service)

	nodePortService := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      service.Name + "-ports",
			Namespace: stack.Namespace,
			Labels:    labels,
		},
		Spec: v1.ServiceSpec{
			Type:     v1.ServiceTypeNodePort,
			Selector: labels,
		},
	}

	for _, con := range containerlist.ForService(service) {
		for i, portBinding := range con.PortBindings {
			if portBinding.Port > 0 {
				continue
			}

			servicePort := v1.ServicePort{
				Name:       fmt.Sprintf("port-%d", i),
				TargetPort: intstr.FromInt(int(portBinding.TargetPort)),
				Port:       int32(portBinding.TargetPort),
			}

			switch portBinding.Protocol {
			case "tcp":
				servicePort.Protocol = v1.ProtocolTCP
			case "udp":
				servicePort.Protocol = v1.ProtocolUDP
			default:
				continue
			}

			nodePortService.Spec.Ports = append(nodePortService.Spec.Ports, servicePort)
		}
	}

	if len(nodePortService.Spec.Ports) > 0 {
		output.Services[nodePortService.Name] = nodePortService
	}
}
