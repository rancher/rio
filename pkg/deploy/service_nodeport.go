package deploy

import (
	"fmt"

	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func nodePorts(objects []runtime.Object, name, namespace string, service *v1beta1.ServiceUnversionedSpec, labels map[string]string) []runtime.Object {
	nodePortService := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: v1.ServiceSpec{
			Type:     v1.ServiceTypeNodePort,
			Selector: labels,
		},
	}

	for i, portBinding := range service.PortBindings {
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

	if len(nodePortService.Spec.Ports) > 0 {
		objects = append(objects, nodePortService)
	}

	return objects
}
