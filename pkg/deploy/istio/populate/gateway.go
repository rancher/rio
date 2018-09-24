package populate

import (
	"github.com/rancher/rio/pkg/deploy/istio/input"
	istioOutput "github.com/rancher/rio/pkg/deploy/istio/output"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func populateGateway(input *input.IstioDeployment, output *istioOutput.Deployment) error {
	if !output.Enabled {
		return nil
	}

	if len(output.Ports) == 0 {
		return nil
	}

	gws := v1alpha3.GatewaySpec{
		Selector: settings.IstioGatewaySelector,
	}

	gateway := &istioOutput.Gateway{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Gateway",
			APIVersion: "networking.istio.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      settings.IstioGateway,
			Namespace: settings.RioSystemNamespace,
		},
		Spec: gws,
	}

	for _, port := range output.Ports {
		gws.Servers = append(gws.Servers, &v1alpha3.Server{
			Port: &v1alpha3.Port{
				Protocol: "HTTP",
				Number:   uint32(port),
			},
			Hosts: []string{"*"},
		})
	}

	output.Gateways[gateway.Name] = gateway
	return nil
}
