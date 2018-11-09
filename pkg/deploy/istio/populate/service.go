package populate

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/rio/pkg/deploy/istio/input"
	"github.com/rancher/rio/pkg/deploy/istio/output"
	"github.com/rancher/rio/pkg/settings"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func getServicePorts(ports []string) []string {
	if len(ports) == 0 {
		return []string{"80/http"}
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
		p, err := strconv.ParseInt(strings.SplitN(port, "/", 2)[0], 10, 0)
		if err != nil {
			return err
		}
		svc.Spec.Ports = append(svc.Spec.Ports, v1.ServicePort{
			Name:       fmt.Sprintf("http-%v", p),
			Protocol:   v1.ProtocolTCP,
			Port:       int32(p),
			TargetPort: intstr.FromInt(int(p)),
		})
	}

	output.Services[svc.Name] = svc
	return nil
}
