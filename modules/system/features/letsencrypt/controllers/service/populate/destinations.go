package populate

import (
	"fmt"
	"hash/adler32"

	"github.com/knative/pkg/apis/istio/v1alpha3"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
)

func DestinationRules(service *v1.Service, ns string, os *objectset.ObjectSet) error {
	for _, dr := range destinationRulesForPublicDomain(service, ns) {
		os.Add(dr)
	}
	return nil
}

func destinationRulesForPublicDomain(service *v1.Service, ns string) []*v1alpha3.DestinationRule {
	var result []*v1alpha3.DestinationRule

	// destinationRule for tls challenge
	for _, publicDomain := range service.Status.PublicDomains {
		checkSum := adler32.Checksum([]byte(publicDomain))
		solverName := fmt.Sprintf("cm-acme-http-solver-%d", checkSum)

		dr := constructors.NewDestinationRule(service.Namespace, solverName, v1alpha3.DestinationRule{
			Spec: v1alpha3.DestinationRuleSpec{
				Host: fmt.Sprintf("%s.%s.svc.cluster.local", solverName, ns),
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
