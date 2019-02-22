package k8sservice

import (
	"fmt"

	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/features/stack/controllers/service/populate/containerlist"
	"github.com/rancher/rio/features/stack/controllers/service/populate/servicelabels"
	"github.com/rancher/rio/pkg/namespace"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	v1client "github.com/rancher/types/apis/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func nodePorts(stack *riov1.Stack, service *riov1.Service, os *objectset.ObjectSet) {
	labels := servicelabels.SelectorLabels(stack, service)

	ns, name := namespace.NameRefWithNamespace(service.Name, stack)
	nodePortService := v1client.NewService(ns, name+"-ports", v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: v1.ServiceSpec{
			Type:     v1.ServiceTypeNodePort,
			Selector: labels,
		},
	})

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
		os.Add(nodePortService)
	}
}
