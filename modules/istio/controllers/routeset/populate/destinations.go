package populate

import (
	"fmt"

	"github.com/knative/pkg/apis/istio/v1alpha3"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func DestinationRules(routeSet *v1.Router, os *objectset.ObjectSet) error {
	os.Add(destinationRuleForRoutes(routeSet)...)
	return nil
}

func destinationRuleForRoutes(route *v1.Router) []runtime.Object {
	var result []runtime.Object

	destSet := populateDestinationFromRouteSet(route)

	for _, dest := range destSet {
		dr := v1alpha3.DestinationRuleSpec{
			Host: fmt.Sprintf("%s.%s.svc.cluster.local", dest.service, dest.stack),
		}

		for revision := range dest.revision {
			dr.Subsets = append(dr.Subsets, v1alpha3.Subset{
				Name: revision,
				Labels: map[string]string{
					"rio.cattle.io/version": revision,
				},
			})
		}

		drObject := newDestinationRuleFromName(dest.service, route.Namespace)
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

func populateDestinationFromRouteSet(routes *v1.Router) map[string]destinationSet {
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

func newDestinationRuleFromName(name, namespace string) *v1alpha3.DestinationRule {
	return constructors.NewDestinationRule(namespace, name, v1alpha3.DestinationRule{
		ObjectMeta: metav1.ObjectMeta{},
	})
}
