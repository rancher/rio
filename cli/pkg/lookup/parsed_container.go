package lookup

import (
	"fmt"

	"github.com/rancher/wrangler/pkg/kv"
)

type ParsedContainer struct {
	PodName       string
	ContainerName string
	Service       StackScoped
}

func ParseContainer(defaultStackName string, name string) (ParsedContainer, bool) {
	result := ParsedContainer{}

	stackScoped := ParseStackScoped(defaultStackName, name)
	if stackScoped.Other == "" {
		return result, false
	}

	result.PodName, result.ContainerName = kv.Split(stackScoped.Other, "/")
	result.Service = stackScoped
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
	return fmt.Sprintf("%s-%s", p.Service.ResourceName, p.PodName)
}
