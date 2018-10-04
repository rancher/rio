package lookup

import (
	"fmt"

	"github.com/rancher/norman/pkg/kv"
	"github.com/rancher/rio/cli/pkg/clientcfg"
	"github.com/rancher/rio/pkg/settings"
)

type ParsedContainer struct {
	PodName       string
	ContainerName string
	Service       StackScoped
}

func ParseContainer(workspace *clientcfg.Workspace, name string) (ParsedContainer, bool) {
	result := ParsedContainer{}

	stackScoped := ParseStackScoped(workspace, name)
	if stackScoped.Other == "" {
		return result, false
	}

	result.PodName, result.ContainerName = kv.Split(stackScoped.Other, "/")
	return result, true
}

func (p ParsedContainer) String() string {
	p.Service.Other = p.PodName
	if p.ContainerName != "" {
		p.Service.Other += "/" + p.ContainerName
	}
	return p.Service.String()
}

func (p ParsedContainer) K8sPodName() string {
	if p.Service.Version == "" || p.Service.Version == settings.DefaultServiceVersion {
		return fmt.Sprintf("%s-%s", p.Service.ResourceName, p.PodName)
	}
	return fmt.Sprintf("%s-%s-%s", p.Service.ResourceName, p.Service.Version, p.PodName)
}
