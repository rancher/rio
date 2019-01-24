package populate

import (
	"fmt"

	"github.com/knative/pkg/apis/istio/v1alpha3"
	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/pkg/namespace"
	v1alpha3client "github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3"
	v1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func DestinationRules(stack *v1.Stack, routeSet *v1.RouteSet, os *objectset.ObjectSet) error {
	os.Add(destinationRuleForRoutes(stack, routeSet)...)
	return nil
}

func destinationRuleForRoutes(stack *v1.Stack, route *v1.RouteSet) []runtime.Object {
	var result []runtime.Object

	destSet := populateDestinationFromRouteSet(route)

	for _, dest := range destSet {
		dr := v1alpha3.DestinationRuleSpec{
			Host: fmt.Sprintf("%s.%s.svc.cluster.local", dest.service, namespace.StackNamespace(stack.Namespace, stack.Name)),
		}

		for revision := range dest.revision {
			dr.Subsets = append(dr.Subsets, v1alpha3.Subset{
				Name: revision,
				Labels: map[string]string{
					"rio.cattle.io/version": revision,
				},
			})
		}

		drObject := newDestinationRuleFromName(stack, dest.service, route.Namespace)
		drObject.Spec = dr
		result = append(result, drObject)
	}

	return result
}

type destinationSet struct {
	service  string
	stack    string
	revision map[string]struct{}
}

func populateDestinationFromRouteSet(routes *v1.RouteSet) map[string]destinationSet {
	result := map[string]destinationSet{}

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

func newDestinationRuleFromName(stack *v1.Stack, name, namespace string) *v1alpha3client.DestinationRule {
	return v1alpha3client.NewDestinationRule(namespace, name, v1alpha3client.DestinationRule{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"rio.cattle.io/stack":   stack.Name,
				"rio.cattle.io/project": stack.Namespace,
			},
		},
	})
}
