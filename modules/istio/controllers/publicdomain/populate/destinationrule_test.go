package populate

import (
	"fmt"
	"hash/adler32"
	"testing"

	"github.com/knative/pkg/apis/istio/v1alpha3"
	"github.com/rancher/rio/modules/test"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
)

func TestPublicDomainDestionationRule(t *testing.T) {
	os := objectset.NewObjectSet()
	pd := adminv1.NewPublicDomain("default", "pd1", adminv1.PublicDomain{
		Spec: adminv1.PublicDomainSpec{
			SecretRef: v1.SecretReference{
				Name:      "pd1-secret",
				Namespace: "default",
			},
			DomainName: "www.foo.com",
		},
	})

	name := fmt.Sprintf("cm-acme-http-solver-%d", adler32.Checksum([]byte(pd.Spec.DomainName)))
	systemNs := "rio-system-fake"
	expected := constructors.NewDestinationRule(systemNs, name, v1alpha3.DestinationRule{
		Spec: v1alpha3.DestinationRuleSpec{
			Host: fmt.Sprintf("%s.%s.svc.cluster.local", name, systemNs),
			Subsets: []v1alpha3.Subset{
				{
					Name: "latest",
				},
			},
		},
	})

	DestionationRule(pd, systemNs, os)

	test.AssertObjects(t, expected, os)
}
