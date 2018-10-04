package istio

import (
	"github.com/rancher/rio/pkg/deploy/istio/input"
	"github.com/rancher/rio/pkg/deploy/istio/output"
	"github.com/rancher/rio/pkg/deploy/istio/populate"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3"
	"k8s.io/api/core/v1"
)

func Remove() error {
	return Deploy(nil, nil, nil)
}

func Deploy(namespace *v1.Namespace, lbService *v1.Service, VirtualServices []*v1alpha3.VirtualService) error {
	input := &input.IstioDeployment{
		LBNamespace:     namespace,
		LBService:       lbService,
		VirtualServices: VirtualServices,
	}
	output := output.NewDeployment()

	if err := populate.Populate(input, output); err != nil {
		return err
	}

	return output.Deploy(settings.RioSystemNamespace, "istio")
}
