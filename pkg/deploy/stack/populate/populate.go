package populate

import (
	"github.com/rancher/rio/pkg/deploy/stack/input"
	"github.com/rancher/rio/pkg/deploy/stack/output"
	"github.com/rancher/rio/pkg/deploy/stack/populate/configmap"
	"github.com/rancher/rio/pkg/deploy/stack/populate/istio"
	"github.com/rancher/rio/pkg/deploy/stack/populate/service"
	"github.com/rancher/rio/pkg/deploy/stack/populate/volume"
)

type populator func(*input.Stack, *output.Deployment) error

func populate(stack *input.Stack, output *output.Deployment, populators ...populator) error {
	for _, populator := range populators {
		if err := populator(stack, output); err != nil {
			return err
		}
	}

	return nil
}

func Populate(stack *input.Stack, output *output.Deployment) error {
	return populate(stack, output,
		configmap.Populate,
		volume.Populate,
		service.Populate,
		istio.Populate)
}
