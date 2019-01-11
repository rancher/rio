package domains

import (
	"fmt"
	"strings"

	"github.com/rancher/rio/api/service"
	"github.com/rancher/rio/pkg/settings"
)

func GetPublicGateway() string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", settings.IstioGateway, settings.RioSystemNamespace)
}

func GetExternalDomain(name, stackName, project string) string {
	parts := strings.Split(project, "-")
	return fmt.Sprintf("%s.%s", service.HashIfNeed(name, strings.SplitN(stackName, "-", 2)[0], parts[len(parts)-1]), settings.ClusterDomain.Get())
}
