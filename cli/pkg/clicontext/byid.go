package clicontext

import (
	"strings"

	"github.com/rancher/rio/cli/pkg/types"
)

func (c *CLIContext) ByID(id string) (types.Resource, error) {
	var t, name, namespace string
	parts := strings.Split(id, "/")
	switch len(parts) {
	case 1:
		t = types.ServiceType
		namespace = c.GetSetNamespace()
		name = parts[0]
	case 2:
		t = parts[0]
		namespace = c.GetSetNamespace()
		name = parts[1]
	case 3:
		t = parts[0]
		namespace = parts[1]
		name = parts[2]
	}

	return c.getResource(types.Resource{
		Namespace: namespace,
		Name:      name,
		Type:      t,
	})
}
