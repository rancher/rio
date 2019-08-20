package populate

import (
	"fmt"
	"hash/adler32"

	"github.com/knative/pkg/apis/istio/v1alpha3"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
)

func DestionationRule(pd *adminv1.PublicDomain, systemNamespace string, os *objectset.ObjectSet) {
	name := fmt.Sprintf("cm-acme-http-solver-%d", adler32.Checksum([]byte(pd.Spec.DomainName)))
	os.Add(constructors.NewDestinationRule(systemNamespace, name, v1alpha3.DestinationRule{
		Spec: v1alpha3.DestinationRuleSpec{
			Host: fmt.Sprintf("%s.%s.svc.cluster.local", name, systemNamespace),
			Subsets: []v1alpha3.Subset{
				{
					Name: "latest",
				},
			},
		},
	}))
	return
}
