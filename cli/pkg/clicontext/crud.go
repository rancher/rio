package clicontext

import (
	"github.com/rancher/mapper"
	"github.com/rancher/rio/cli/pkg/types"
	"k8s.io/apimachinery/pkg/runtime"
)

func (c *CLIContext) MultiCreate(objs ...runtime.Object) error {
	var (
		errors []error
	)

	for _, obj := range objs {
		err := c.Create(obj)
		errors = append(errors, err)
	}

	return mapper.NewErrors(errors...)
}

func (c *CLIContext) UpdateResource(r types.Resource, updater func(obj runtime.Object) error) error {
	r, err := c.getResource(r)
	if err != nil {
		return err
	}
	if err := updater(r.Object); err != nil {
		return err
	}

	return c.UpdateObject(r.Object)
}

func (c *CLIContext) Update(name, typeName string, updater func(obj runtime.Object) error) error {
	r, err := c.ByID(name)
	if err != nil {
		return err
	}
	return c.UpdateResource(r, updater)
}
