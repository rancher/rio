package domains

import (
	"fmt"

	"github.com/rancher/rio/pkg/settings"
)

func GetPublicGateway(systemNamespace string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", settings.RioGateway, systemNamespace)
}

func GetExternalDomain(name, namespace, clusterDomain string) string {
	return fmt.Sprintf("%s-%s.%s", name, namespace, clusterDomain)
}
