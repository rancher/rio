package k8sservice

import (
	"fmt"

	"github.com/rancher/rio/modules/stack/controllers/service/populate/containerlist"
	"github.com/rancher/rio/modules/stack/controllers/service/populate/servicelabels"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func nodePorts(stack *riov1.Stack, service *riov1.Service, os *objectset.ObjectSet) {
	labels := servicelabels.SelectorLabels(stack, service)

	nodePortService := constructors.NewService(stack.Name, service.Name+"-ports", v1.Service{
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
