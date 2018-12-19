package populate

import (
	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/pkg/settings"
	projectv1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	"k8s.io/api/core/v1"
)

func Istio(publicdomains []*projectv1.PublicDomain, secret *v1.Secret) *objectset.ObjectSet {
	output := objectset.NewObjectSet()
	if settings.IstioEnabled.Get() != "true" {
		return output
	}

	output.AddInput(secret)
	for _, pd := range publicdomains {
		output.AddInput(pd)
	}

	if err := populateStack(output); err != nil {
		output.AddErr(err)
	}

	populateGateway(secret, publicdomains, output)

	return output
}
