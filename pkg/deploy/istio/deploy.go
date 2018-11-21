package istio

import (
	"github.com/rancher/rio/pkg/deploy/istio/input"
	"github.com/rancher/rio/pkg/deploy/istio/output"
	"github.com/rancher/rio/pkg/deploy/istio/populate"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3"
	"github.com/rancher/rio/types/apis/space.cattle.io/v1beta1"
	"k8s.io/api/core/v1"
)

func Remove() error {
	return Deploy(nil, nil, nil, nil)
}

func Deploy(namespace *v1.Namespace, VirtualServices []*v1alpha3.VirtualService, publicdomains []*v1beta1.PublicDomain, secret *v1.Secret) error {
	input := &input.IstioDeployment{
		LBNamespace:     namespace,
		VirtualServices: VirtualServices,
		PublicDomains:   publicdomains,
		Secret:          secret,
	}
	output := output.NewDeployment()

	if err := populate.Populate(input, output); err != nil {
		return err
	}

	return output.Deploy(settings.RioSystemNamespace, "istio")
}
