package populate

import (
	"fmt"

	"github.com/rancher/rio/pkg/deploy/istio/input"
	"github.com/rancher/rio/pkg/deploy/istio/output"
	"github.com/rancher/rio/pkg/settings"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func getServicePorts(ports []int) []int {
	if len(ports) == 0 {
		return []int{80}
	}
	return ports
}

func populateService(input *input.IstioDeployment, output *output.Deployment) error {
	if !output.Enabled || input.LBNamespace == nil {
		return nil
	}

	svc := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      settings.IstioExternalLB,
			Namespace: input.LBNamespace.Name,
		},
		Spec: v1.ServiceSpec{
			Type:     v1.ServiceTypeLoadBalancer,
			Selector: settings.IstioGatewaySelector,
		},
	}

	for _, port := range getServicePorts(output.Ports) {
		svc.Spec.Ports = append(svc.Spec.Ports, v1.ServicePort{
			Name:       fmt.Sprintf("http-%d", port),
			Protocol:   v1.ProtocolTCP,
			Port:       int32(port),
			TargetPort: intstr.FromInt(int(port)),
		})
	}

	output.Services[svc.Name] = svc
	return nil
}
