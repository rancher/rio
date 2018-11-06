package populate

import (
	"strings"

	"github.com/rancher/rio/pkg/deploy/istio/input"
	"github.com/rancher/rio/pkg/deploy/istio/output"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3"
)

func populatePorts(input *input.IstioDeployment, output *output.Deployment) error {
	ports := sortedPorts(input.VirtualServices)

	output.Enabled = settings.IstioEnabled.Get() == "true"
	output.Ports = ports
	output.UseLoadBalancer = !isHostPorts(input)

	return nil
}

func isHostPorts(input *input.IstioDeployment) bool {
	if input.LBService != nil {
		for _, ingress := range input.LBService.Status.LoadBalancer.Ingress {
			if ingress.Hostname == "localhost" || ingress.IP != "" {
				return false
			}
		}
	}

	return false
}

func sortedPorts(vss []*v1alpha3.VirtualService) []string {
	var result []string
	for _, vs := range vss {
		for _, port := range getPorts(vs) {
			result = append(result, port)
		}
	}

	return result
}

func getPorts(service *v1alpha3.VirtualService) []string {
	ports, ok := service.Annotations["rio.cattle.io/ports"]
	if !ok || ports == "" {
		return nil
	}

	return strings.Split(ports, ",")
}
