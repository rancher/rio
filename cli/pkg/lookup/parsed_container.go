package lookup

import (
	"strings"

	"github.com/rancher/wrangler/pkg/kv"
)

type ParsedContainer struct {
	PodName       string
	ContainerName string
	Service       StackScoped
}

func ParseContainer(defaultStackName string, name string) (ParsedContainer, bool) {
	result := ParsedContainer{}

	var stackScoped StackScoped
	if len(strings.Split(name, "/")) == 4 {
		stackScoped = ParseStackScoped(defaultStackName, name)
	} else {
		namespace, other := kv.Split(name, "/")
		stackScoped = StackScoped{
			StackName: namespace,
			Other:     other,
		}
	}

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
