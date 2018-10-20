package istio

import (
	"github.com/rancher/rio/pkg/deploy/stack/input"
	"github.com/rancher/rio/pkg/deploy/stack/output"
	"github.com/rancher/rio/pkg/deploy/stack/populate/service"
	"github.com/rancher/rio/pkg/deploy/stack/populate/servicelabels"
	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"istio.io/api/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func destinationRules(stack *input.Stack) ([]*output.IstioObject, error) {
	var result []*output.IstioObject
	services, err := service.CollectionServices(stack.Services)
	if err != nil {
		return nil, err
	}

	for name, service := range services {
		result = append(result, append(destinationRuleForService(stack, name, service), destinationRuleForRevisionService(stack, name, service)...)...)
	}

	for _, routeset := range stack.RouteSet {
		result = append(result, destinationRuleForRoutes(stack, routeset)...)
	}

	return result, nil
}

func destinationRuleForRoutes(stack *input.Stack, route *v1beta1.RouteSet) []*output.IstioObject {
	result := make([]*output.IstioObject, 0)
	destSet := populateDestinationFromRouteset(route)
	ns := namespace.StackNamespace(stack.Stack.Namespace, stack.Stack.Name)
	for _, dest := range destSet {
		dr := v1alpha3.DestinationRule{
			Host: dest.service,
		}
		for revision := range dest.revision {
			dr.Subsets = append(dr.Subsets, &v1alpha3.Subset{
				Name: revision,
				Labels: map[string]string{
					"rio.cattle.io/version": revision,
				},
			})
		}
		drObject := newDestinationRuleFromName(stack, dest.service, ns)
		drObject.Spec = &dr
		result = append(result, drObject)
	}
	return result
}

type destinationSet struct {
	service  string
	stack    string
	revision map[string]struct{}
}

func populateDestinationFromRouteset(routes *v1beta1.RouteSet) map[string]destinationSet {
	result := make(map[string]destinationSet, 0)
	for _, spec := range routes.Spec.Routes {
		for _, dest := range spec.To {
			key := dest.Stack + "-" + dest.Service
			if _, ok := result[key]; !ok {
				result[key] = destinationSet{
					service:  dest.Service,
					stack:    dest.Stack,
					revision: map[string]struct{}{},
				}
			}
			if dest.Revision == "" {
				dest.Revision = "v0"
			}
			result[key].revision[dest.Revision] = struct{}{}
		}
	}
	return result
}

func destinationRuleForService(stack *input.Stack, name string, service *output.ServiceSet) []*output.IstioObject {
	result := make([]*output.IstioObject, 0)
	dr := v1alpha3.DestinationRule{
		Host: name,
	}

	dr.Subsets = append(dr.Subsets, newSubSet(service.Service))

	for _, rev := range service.Revisions {
		dr.Subsets = append(dr.Subsets, newSubSet(rev))
	}

	drObject := newDestinationRule(stack, service.Service)
	drObject.Spec = &dr
	result = append(result, drObject)

	return result
}

func destinationRuleForRevisionService(stack *input.Stack, name string, service *output.ServiceSet) []*output.IstioObject {
	result := make([]*output.IstioObject, 0)
	for _, rev := range service.Revisions {
		dr := v1alpha3.DestinationRule{
			Host: rev.Name,
		}
		dr.Subsets = append(dr.Subsets, newSubSet(rev))
		drObject := newDestinationRule(stack, rev)
		drObject.Spec = &dr
		result = append(result, drObject)
	}
	return result

}

func newSubSet(service *v1beta1.Service) *v1alpha3.Subset {
	return &v1alpha3.Subset{
		Name: service.Spec.Revision.Version,
		Labels: map[string]string{
			"rio.cattle.io/version": service.Spec.Revision.Version,
		},
	}
}

func newDestinationRule(stack *input.Stack, service *v1beta1.Service) *output.IstioObject {
	labels := servicelabels.RioOnlyServiceLabels(stack, service)
	return &output.IstioObject{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "DestinationRule",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      service.Name,
			Namespace: service.Namespace,
			Labels:    labels,
		},
	}
}

func newDestinationRuleFromName(stack *input.Stack, name, namespace string) *output.IstioObject {
	return &output.IstioObject{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "DestinationRule",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"rio.cattle.io/stack":     stack.Stack.Name,
				"rio.cattle.io/workspace": stack.Stack.Namespace,
			},
		},
	}
}
