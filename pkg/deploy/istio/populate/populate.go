package populate

import (
	"github.com/rancher/rio/pkg/deploy/istio/input"
	"github.com/rancher/rio/pkg/deploy/istio/output"
)

type populator func(*input.IstioDeployment, *output.Deployment) error

func populate(stack *input.IstioDeployment, output *output.Deployment, populators ...populator) error {
	for _, populator := range populators {
		if err := populator(stack, output); err != nil {
			return err
		}
	}

	return nil
}

func Populate(stack *input.IstioDeployment, output *output.Deployment) error {
	return populate(stack, output,
		populatePorts,
		populateStack,
		populateService,
		populateGateway)
}
