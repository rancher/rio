package populate

import (
	"fmt"

	"github.com/knative/pkg/apis/istio/v1alpha3"
	projectv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
)

func DestinationRulesAndVirtualServices(namespace string, clusterDomain *projectv1.ClusterDomain, service *v1.Service, os *objectset.ObjectSet) error {
	return virtualServices(namespace, clusterDomain, service, os)
}

func DestinationRuleForService(app *riov1.App) *v1alpha3.DestinationRule {
	drSpec := v1alpha3.DestinationRuleSpec{
		Host: fmt.Sprintf("%s.%s.svc.cluster.local", app.Name, app.Namespace),
	}

	for _, rev := range app.Spec.Revisions {
		drSpec.Subsets = append(drSpec.Subsets, newSubSet(rev.Version))
	}

	dr := newDestinationRule(app.Namespace, app.Name)
	dr.Spec = drSpec

	return dr
}

func newSubSet(version string) v1alpha3.Subset {
	return v1alpha3.Subset{
		Name: version,
		Labels: map[string]string{
			"version": version,
		},
	}
}

func newDestinationRule(namespace, name string) *v1alpha3.DestinationRule {
	return constructors.NewDestinationRule(namespace, name, v1alpha3.DestinationRule{})
}
