package domains

import (
	"fmt"

	"github.com/rancher/rio/pkg/constants"
)

func GetPublicGateway(systemNamespace string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", constants.RioGateway, systemNamespace)
}

func GetExternalDomain(name, namespace, clusterDomain string) string {
	return fmt.Sprintf("%s-%s.%s", name, namespace, clusterDomain)
}
