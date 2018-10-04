package clicontext

import (
	"fmt"
	"strings"

	"github.com/rancher/norman/types"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	spaceclient "github.com/rancher/rio/types/client/space/v1beta1"
)

func (c *CLIContext) ByID(id, typeName string) (*types.NamedResource, error) {
	switch typeName {
	case spaceclient.PodType:
		return c.podById(id, typeName)
	case client.ServiceType:
		fallthrough
	case client.VolumeType:
		fallthrough
	case client.ConfigType:
		return c.stackScopedById(id, typeName)
	}

	return c.defaultById(id, typeName)
}

func (c *CLIContext) podById(id, schemaType string) (*types.NamedResource, error) {
	parts := strings.Split(id, "/")
	if len(parts) > 2 || !strings.Contains(parts[0], ":") {
		return nil, nil
	}

	return c.defaultById(parts[0], schemaType)
}

func (c *CLIContext) stackScopedById(id, schemaType string) (*types.NamedResource, error) {
	result, err := c.defaultById(id, schemaType)
	if err != nil || result != nil {
		return result, err
	}

	w, err := c.Workspace()
	if err != nil {
		return nil, err
	}

	scoped := lookup.ParseStackScoped(w, id)
	return c.defaultById(scoped.ResourceID, schemaType)
}

func (c *CLIContext) defaultById(id, schemaType string) (*types.NamedResource, error) {
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
