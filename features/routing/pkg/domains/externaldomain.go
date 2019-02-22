package domains

import (
	"fmt"
	"strings"

	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/pkg/settings"
)

func GetPublicGateway() string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", settings.IstioGateway, settings.RioSystemNamespace)
}

func GetExternalDomain(name, stackName, projectId string) string {
	return fmt.Sprintf("%s.%s", namespace.HashIfNeed(name, strings.SplitN(stackName, "-", 2)[0], projectId), settings.ClusterDomain.Get())
}
