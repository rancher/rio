package clicontext

import (
	"fmt"
	"strings"

	"github.com/rancher/norman/pkg/kv"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/types"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *CLIContext) ByID(id, typeName string) (types.Resource, error) {
	var r types.Resource
	var err error
	switch typeName {
	case types.PodType:
		r, err = c.podByID(id, typeName)
	case types.ServiceType:
		fallthrough
	case types.VolumeType:
		fallthrough
	case types.ConfigType:
		r, err = c.stackScopedByID(id, typeName)
	default:
		r, err = c.defaultByID(id, typeName)
	}
	if err != nil {
		return r, err
	}

	return r, c.getResource(r)
}

func (c *CLIContext) getResource(resource types.Resource) error {
	client, err := c.KubeClient()
	if err != nil {
		return err
	}

	switch resource.Type {
	case clitypes.ServiceType:
		if _, err := client.Rio.Services(resource.Namespace).Get(resource.Name, metav1.GetOptions{}); err != nil {
			return err
		}
	case clitypes.StackType:
		if _, err := client.Rio.Stacks(resource.Namespace).Get(resource.Name, metav1.GetOptions{}); err != nil {
			return err
		}
	case clitypes.PodType:
		podname, _ := kv.Split(resource.Name, "/")
		if _, err := client.Core.Pods(resource.Namespace).Get(podname, metav1.GetOptions{}); err != nil {
			return err
		}
	case clitypes.ConfigType:
		if _, err := client.Rio.Configs(resource.Namespace).Get(resource.Name, metav1.GetOptions{}); err != nil {
			return err
		}
	case clitypes.RouteSetType:
		if _, err := client.Rio.Services(resource.Namespace).Get(resource.Name, metav1.GetOptions{}); err != nil {
			return err
		}
	case clitypes.VolumeType:
		if _, err := client.Rio.Volumes(resource.Namespace).Get(resource.Name, metav1.GetOptions{}); err != nil {
			return err
		}
	case clitypes.ExternalServiceType:
		if _, err := client.Rio.ExternalServices(resource.Namespace).Get(resource.Name, metav1.GetOptions{}); err != nil {
			return err
		}
	}
	return nil
}

func (c *CLIContext) podByID(id, schemaType string) (types.Resource, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 4 {
		return types.Resource{}, nil
	}
	return types.Resource{
		Name:      fmt.Sprintf("%s/%s", strings.Join(parts[1:3], "-"), parts[3]),
		Namespace: parts[0],
		Type:      schemaType,
	}, nil
}

func (c *CLIContext) stackScopedByID(id, schemaType string) (types.Resource, error) {
	w, err := c.Project()
	if err != nil {
		return types.Resource{}, err
	}

	scoped := lookup.ParseStackScoped(w, id)
	if scoped.Other != "" {
		return types.Resource{}, fmt.Errorf("invalid stack scoped ID")
	}

	return types.Resource{
		Type:      schemaType,
		Name:      scoped.ResourceName,
		Namespace: scoped.StackName,
	}, nil
}

func (c *CLIContext) defaultByID(id, schemaType string) (types.Resource, error) {
	if !strings.Contains(id, ":") || strings.Contains(id, "/") {
		return types.Resource{}, fmt.Errorf("invalid id format")
	}

	name, namespace := kv.Split(id, "/")

	return types.Resource{
		Namespace: namespace,
		Name:      name,
		Type:      schemaType,
	}, nil
}
