package clicontext

import (
	"fmt"
	"strings"

	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/pkg/namespace"

	"github.com/rancher/rio/types/client/rio/v1beta1"
	spaceclient "github.com/rancher/rio/types/client/space/v1beta1"
)

func (c *CLIContext) LookupFilters(name, typeName string) (map[string]interface{}, bool, error) {
	filters := map[string]interface{}{
		"name":         name,
		"removed_null": "1",
	}

	switch typeName {
	case client.StackType:
		return filters, true, nil
	case client.RouteSetType:
		return c.routeSetByName(filters, name)
	case spaceclient.PodType:
		return c.podByName(filters, name)
	default:
		return c.defaultByName(filters, name)
	}
}

func (c *CLIContext) defaultByName(filters map[string]interface{}, name string) (map[string]interface{}, bool, error) {
	w, err := c.Workspace()
	if err != nil {
		return nil, false, err
	}

	stackScoped := lookup.ParseStackScoped(w, name)
	return c.stackScopedByName(filters, stackScoped.StackName, stackScoped.LookupName())
}

func (c *CLIContext) routeSetByName(filters map[string]interface{}, name string) (map[string]interface{}, bool, error) {
	var (
		stackName   string
		serviceName string
	)
	parts := strings.SplitN(name, "://", 2)
	if len(parts) > 1 {
		parts[0] = parts[1]
	}
	parts = strings.Split(parts[0], ".")
	if len(parts) == 1 {
		stackName = "default"
		serviceName = parts[0]
	} else {
		stackName = parts[1]
		serviceName = parts[0]
	}

	return c.stackScopedByName(filters, stackName, serviceName)
}

func (c *CLIContext) podByName(filters map[string]interface{}, name string) (map[string]interface{}, bool, error) {
	w, err := c.Workspace()
	if err != nil {
		return nil, false, err
	}

	pc, ok := lookup.ParseContainer(w, name)
	if !ok {
		return nil, false, nil
	}

	filters["name"] = pc.K8sPodName()
	filters["namespace"] = namespace.StackNamespace(w.ID, pc.Service.StackName)
	return filters, true, nil
}

func (c *CLIContext) stackScopedByName(filters map[string]interface{}, stackName, resourceName string) (map[string]interface{}, bool, error) {
	w, err := c.Workspace()
	if err != nil {
		return nil, false, err
	}
	stackID := fmt.Sprintf("%s:%s", w.ID, stackName)
	filters["stackId"] = stackID
	filters["name"] = resourceName
	return filters, true, nil
}
