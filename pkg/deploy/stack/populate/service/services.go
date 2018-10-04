package service

import (
	"github.com/rancher/rio/pkg/deploy/stack/input"
	"github.com/rancher/rio/pkg/deploy/stack/output"
	"github.com/rancher/rio/pkg/deploy/stack/populate/k8sservice"
	"github.com/rancher/rio/pkg/deploy/stack/populate/podcontrollers"
)

func Populate(stack *input.Stack, output *output.Deployment) error {
	var err error

	serviceSet, err := CollectionServices(stack.Services)
	if err != nil {
		return err
	}

	for _, s := range serviceSet.List() {
		k8sservice.Populate(stack, s, output)
		if err := podcontrollers.Populate(stack, s, output); err != nil {
			return err
		}
	}

	return nil
}
