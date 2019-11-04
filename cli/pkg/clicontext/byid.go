package clicontext

import (
	"github.com/rancher/rio/cli/pkg/types"
	"github.com/rancher/wrangler/pkg/kv"
)

func (c *CLIContext) ParseIDForType(id, resourceType string) types.Resource {
	r := c.ParseID(id)
	r.Type = resourceType
	if r.Type != types.ServiceType {
		r.App = ""
		r.Version = ""
	}
	return r
}

func (c *CLIContext) ParseID(id string) types.Resource {
	namespace, typeAndName := kv.RSplit(id, ":")
	t, nameAndVersion := kv.RSplit(typeAndName, "/")
	name, version := kv.Split(nameAndVersion, "@")

	if namespace == "" {
		namespace = c.GetSetNamespace()
	}

	if t == "" {
		t = types.ServiceType
	} else if normalize, ok := types.Aliases[t]; ok {
		t = normalize
	}

	if version == "" && t == types.ServiceType {
		version = types.DefaultVersion
	}

	return types.Resource{
		LookupName: id,
		Namespace:  namespace,
		Name:       nameAndVersion,
		Type:       t,
		App:        name,
		Version:    version,
	}
}

func (c *CLIContext) ByID(id string) (types.Resource, error) {
	return c.getResource(c.ParseID(id))
}
