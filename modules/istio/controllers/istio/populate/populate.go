package populate

import (
	projectv1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
)

func Istio(systemNamespace string, clusterDomain *projectv1.ClusterDomain, publicdomains []*riov1.PublicDomain, wildcardSecret *v1.Secret) *objectset.ObjectSet {
	output := objectset.NewObjectSet()
	populateGateway(systemNamespace, clusterDomain, wildcardSecret, publicdomains, output)

	return output
}
