package populate

import (
	"fmt"
	"hash/adler32"
	"strings"

	"github.com/knative/pkg/apis/istio/v1alpha3"
	"github.com/rancher/rio/features/routing/controllers/service/populate"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
)

func DestinationRules(service *v1.Service, os *objectset.ObjectSet) error {
	for _, dr := range destinationRulesForPublicDomain(service) {
		os.Add(dr)
	}
	return nil
}

func destinationRulesForPublicDomain(service *v1.Service) []*v1alpha3.DestinationRule {
	var result []*v1alpha3.DestinationRule

	// destinationRule for tls challenge
	for _, publicDomain := range strings.Split(service.Annotations[populate.PublicDomainAnnotation], ",") {
		if publicDomain == "" {
			continue
		}

		checkSum := adler32.Checksum([]byte(publicDomain))
		solverName := fmt.Sprintf("cm-acme-http-solver-%d", checkSum)

		dr := constructors.NewDestinationRule(service.Namespace, solverName, v1alpha3.DestinationRule{
			Spec: v1alpha3.DestinationRuleSpec{
				Host: fmt.Sprintf("%s.rio-system.svc.cluster.local", solverName),
				Subsets: []v1alpha3.Subset{
					{
						Name: "latest",
						Labels: map[string]string{
							"certmanager.k8s.io/acme-http-domain": fmt.Sprint(checkSum),
						},
					},
				},
			},
		})

		result = append(result, dr)
	}

	return result
}
