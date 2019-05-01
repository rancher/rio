package clicontext

import (
	"github.com/rancher/rio/cli/pkg/types"
)

func (c *CLIContext) ByID(namespace, name string, typeName string) (types.Resource, error) {
	return c.getResource(types.Resource{
		Namespace: namespace,
		Name:      name,
		Type:      typeName,
	})
}
