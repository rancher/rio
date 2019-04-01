package populate

import (
	projectv1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
)

func Istio(systemNamespace string, clusterDomain *projectv1.ClusterDomain, publicdomains []*projectv1.PublicDomain, secret *v1.Secret) *objectset.ObjectSet {
	output := objectset.NewObjectSet()
	populateGateway(systemNamespace, clusterDomain, secret, publicdomains, output)

	return output
}
