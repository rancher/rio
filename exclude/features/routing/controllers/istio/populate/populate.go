package populate

import (
	projectv1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
)

func Istio(systemNamespace string, publicdomains []*projectv1.PublicDomain, secret *v1.Secret) *objectset.ObjectSet {
	output := objectset.NewObjectSet()
	output.AddInput(secret)
	for _, pd := range publicdomains {
		output.AddInput(pd)
	}

	populateGateway(systemNamespace, secret, publicdomains, output)

	return output
}
