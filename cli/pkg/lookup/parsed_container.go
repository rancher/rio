package lookup

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	containerNameRegexp = regexp.MustCompile("(.*)/([a-f0-9]+-[a-z0-9]{5})(/[^/]+)?")
)

type ParsedContainer struct {
	PodName       string
	ContainerName string
	Service       ParsedService
}

func ParseContainerName(name string) (ParsedContainer, bool) {
	var result ParsedContainer
	matches := containerNameRegexp.FindStringSubmatch(name)
	if matches == nil {
		return result, false
	}

	result.Service = ParseServiceName(matches[1])
	result.ContainerName = matches[3]
	if result.ContainerName == "" {
		result.ContainerName = result.Service.ServiceName
	} else {
		result.ContainerName = result.ContainerName[1:]
	}
	result.PodName = result.Service.PodNamePrefix() + matches[2]

	return result, true
}

func (p ParsedContainer) String() string {
	name := fmt.Sprintf(
		"%s/%s/%s",
		p.Service.String(),
		strings.TrimPrefix(p.PodName, p.Service.PodNamePrefix()),
		p.ContainerName)
	return strings.TrimSuffix(name, "/"+p.Service.ServiceName)
}
