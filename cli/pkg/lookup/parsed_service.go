package lookup

import (
	"fmt"

	"github.com/rancher/rio/cli/pkg/constants"

	"github.com/rancher/rio/exclude/pkg/settings"
	"github.com/rancher/wrangler/pkg/kv"
)

type StackScoped struct {
	DefaultStackName string
	StackName        string
	ResourceName     string
	Version          string
	ServiceName      string
	Other            string
}

func StackScopedFromLabels(defaultStackName string, labels map[string]string) StackScoped {
	return StackScoped{
		DefaultStackName: defaultStackName,
		Version:          labels["rio.cattle.io/version"],
		StackName:        labels["rio.cattle.io/stack"],
		ResourceName:     labels["rio.cattle.io/service-name"],
		ServiceName:      labels["rio.cattle.io/service"],
	}
}

func ParseStackScoped(defaultStackName string, serviceName string) StackScoped {
	var result StackScoped
	result.StackName, result.ResourceName = kv.Split(serviceName, "/")
	if result.ResourceName == "" {
		result.ResourceName = result.StackName
		result.StackName = defaultStackName
	}
	result.ResourceName, result.Other = kv.Split(result.ResourceName, "/")
	result.ResourceName, result.Version = kv.Split(result.ResourceName, ":")

	if result.Version == "" || result.Version == constants.DefaultServiceVersion {
		result.ServiceName = result.ResourceName
	} else {
		result.ServiceName = result.ResourceName
		result.ResourceName = fmt.Sprintf("%s-%s", result.ResourceName, result.Version)
	}
	return result
}

func (p StackScoped) String() string {
	result := ""

	if p.StackName != "" {
		if p.Other != "" || p.StackName != p.DefaultStackName {
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
