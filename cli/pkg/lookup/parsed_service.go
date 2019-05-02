package lookup

import (
	"fmt"

	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/wrangler/pkg/kv"
	corev1 "k8s.io/api/core/v1"
)

type StackScoped struct {
	DefaultStackName string
	StackName        string
	ResourceName     string
	Version          string
	ServiceName      string
	Other            string
}

func StackScopedFromLabels(defaultStackName string, pod *corev1.Pod) StackScoped {
	return StackScoped{
		DefaultStackName: defaultStackName,
		Version:          pod.Labels["version"],
		StackName:        pod.Namespace,
		ServiceName:      pod.Labels["app"],
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

	result.ServiceName = result.ResourceName
	if result.Version != "" {
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
