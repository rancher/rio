package domains

import (
	"fmt"
	"strings"

	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/pkg/settings"
)

func GetPublicGateway(systemNamespace string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", settings.RioGateway, systemNamespace)
}

func GetExternalDomain(name, stackName, projectName string) string {
	return fmt.Sprintf("%s.%s", namespace.HashIfNeed(name, strings.SplitN(stackName, "-", 2)[0], projectName), settings.ClusterDomain)
}
