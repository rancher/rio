package lookup

import (
	"fmt"

	"github.com/rancher/norman/pkg/kv"
	"github.com/rancher/rio/cli/pkg/clientcfg"
	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/pkg/settings"
)

type StackScoped struct {
	Workspace    *clientcfg.Workspace
	StackName    string
	ResourceName string
	ResourceID   string
	Version      string
	Other        string
}

func StackScopedFromLabels(workspace *clientcfg.Workspace, labels map[string]string) StackScoped {
	service := labels["rio.cattle.io/service"]
	serviceName := labels["rio.cattle.io/service-name"]
	stack := labels["rio.cattle.io/stack"]
	rev := labels["rio.cattle.io/version"]

	ns := namespace.StackNamespace(workspace.ID, stack)

	return StackScoped{
		Workspace:    workspace,
		Version:      rev,
		StackName:    stack,
		ResourceID:   fmt.Sprintf("%s:%s", ns, serviceName),
		ResourceName: service,
	}
}

func ParseStackScoped(workspace *clientcfg.Workspace, serviceName string) StackScoped {
	var result StackScoped
	result.StackName, result.ResourceName = kv.Split(serviceName, "/")
	if result.ResourceName == "" {
		result.ResourceName = result.StackName
		result.StackName = workspace.Cluster.DefaultStackName
	}
	result.ResourceName, result.Other = kv.Split(result.ResourceName, "/")
	result.ResourceName, result.Version = kv.Split(result.ResourceName, ":")
	result.Workspace = workspace

	name := fmt.Sprintf("%s-%s", result.ResourceName, result.Version)
	if result.Version == "" || result.Version == settings.DefaultServiceVersion {
		name = result.ResourceName
	}
	result.ResourceID = fmt.Sprintf("%s:%s", namespace.StackNamespace(workspace.ID, result.StackName), name)
	return result
}

func (p StackScoped) LookupName() string {
	if p.Version == "" || p.Version == settings.DefaultServiceVersion {
		return p.ResourceName
	}
	return fmt.Sprintf("%s-%s", p.ResourceName, p.Version)
}

func (p StackScoped) String() string {
	result := ""

	if p.StackName != "" {
		if p.Other != "" || p.StackName != p.Workspace.Cluster.DefaultStackName {
			result = p.StackName + "/"
		}
	}

	result += p.ResourceName

	if p.Version != "" && p.Version != settings.DefaultServiceVersion {
		result += ":" + p.Version
	}

	if p.Other != "" {
		result += "/" + p.Other
	}

	return result
}
