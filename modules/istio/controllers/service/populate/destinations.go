package populate

import (
	"fmt"

	"github.com/knative/pkg/apis/istio/v1alpha3"
	projectv1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/rio/pkg/serviceset"
	"github.com/rancher/wrangler/pkg/objectset"
)

func DestinationRulesAndVirtualServices(namespace string, clusterDomain *projectv1.ClusterDomain, serviceSet *serviceset.ServiceSet, service *v1.Service, os *objectset.ObjectSet) error {
	if err := destinationRules(service, serviceSet, os); err != nil {
		return err
	}
	return virtualServices(namespace, clusterDomain, serviceSet, service, os)
}

func destinationRules(service *v1.Service, serviceSet *serviceset.ServiceSet, os *objectset.ObjectSet) error {
	dr := destinationRuleForService(service.Namespace, service.Name, serviceSet)
	os.Add(dr)

	for _, dr := range destinationRuleForRevisionService(serviceSet) {
		os.Add(dr)
	}

	return nil
}

func destinationRuleForService(namespace, name string, service *serviceset.ServiceSet) *v1alpha3.DestinationRule {
	drSpec := v1alpha3.DestinationRuleSpec{
		Host: fmt.Sprintf("%s.%s.svc.cluster.local", name, namespace),
	}

	for _, rev := range service.Revisions {
		drSpec.Subsets = append(drSpec.Subsets, newSubSet(rev))
	}

	dr := newDestinationRule(namespace, name)
	dr.Spec = drSpec

	return dr
}

func destinationRuleForRevisionService(service *serviceset.ServiceSet) []*v1alpha3.DestinationRule {
	var result []*v1alpha3.DestinationRule

	for _, rev := range service.Revisions {
		drObject := newDestinationRule(rev.Namespace, rev.Name)
		drObject.Spec = v1alpha3.DestinationRuleSpec{
			Host: rev.Name,
			Subsets: []v1alpha3.Subset{
				newSubSet(rev),
			},
		}
		result = append(result, drObject)
	}

	return result
}

func newSubSet(service *v1.Service) v1alpha3.Subset {
	_, version := services.AppAndVersion(service)
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
