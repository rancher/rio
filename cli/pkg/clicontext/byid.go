package clicontext

import (
	"fmt"
	"strings"

	"github.com/rancher/norman/types"
	"github.com/rancher/rio/cli/pkg/lookup"
	projectclient "github.com/rancher/rio/types/client/project/v1"
	"github.com/rancher/rio/types/client/rio/v1"
)

func (c *CLIContext) ByID(id, typeName string) (*types.NamedResource, error) {
	switch typeName {
	case projectclient.PodType:
		return c.podByID(id, typeName)
	case client.ServiceType:
		fallthrough
	case client.VolumeType:
		fallthrough
	case client.ConfigType:
		return c.stackScopedByID(id, typeName)
	}

	return c.defaultByID(id, typeName)
}

func (c *CLIContext) podByID(id, schemaType string) (*types.NamedResource, error) {
	parts := strings.Split(id, "/")
	if len(parts) > 2 || !strings.Contains(parts[0], ":") {
		return nil, nil
	}

	return c.defaultByID(parts[0], schemaType)
}

func (c *CLIContext) stackScopedByID(id, schemaType string) (*types.NamedResource, error) {
	result, err := c.defaultByID(id, schemaType)
	if err == nil {
		return result, err
	}

	w, err := c.Project()
	if err != nil {
		return nil, err
	}

	scoped := lookup.ParseStackScoped(w, id)
	if scoped.Other != "" {
		return nil, fmt.Errorf("invalid stack scoped ID")
	}

	return c.defaultByID(scoped.ResourceID, schemaType)
}

func (c *CLIContext) defaultByID(id, schemaType string) (*types.NamedResource, error) {
	var resource types.NamedResource

	if !strings.Contains(id, ":") || strings.Contains(id, "/") {
		return nil, fmt.Errorf("invalid id format")
	}

	client, err := c.ClientLookup(schemaType)
	if err != nil {
		return nil, err
	}

	err = client.ByID(schemaType, id, &resource)
	return &resource, err
}
