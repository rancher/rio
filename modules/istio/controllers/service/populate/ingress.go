package populate

import (
	"fmt"

	name2 "github.com/rancher/rio/pkg/name"

	"github.com/rancher/rio/modules/istio/pkg/domains"

	"github.com/rancher/rio/modules/system/features/letsencrypt/pkg/issuers"
	projectv1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/wrangler/pkg/objectset"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Ingress(clusterDomain *projectv1.ClusterDomain, systemNamespace string, ns, name string, tls bool, os *objectset.ObjectSet) error {
	if clusterDomain.Status.ClusterDomain == "" {
		return nil
	}
	ingress := constructors.NewIngress(systemNamespace, name2.SafeConcatName(name, ns), v1beta1.Ingress{})

	wildcardsDomain := fmt.Sprintf("*.%s", clusterDomain.Status.ClusterDomain)
	domain := domains.GetExternalDomain(name, ns, clusterDomain.Status.ClusterDomain)
	ingress.Spec.Rules = []v1beta1.IngressRule{
		{
			Host: domain,
			IngressRuleValue: v1beta1.IngressRuleValue{
				HTTP: &v1beta1.HTTPIngressRuleValue{
					Paths: []v1beta1.HTTPIngressPath{
						{
							Path: "/",
							Backend: v1beta1.IngressBackend{
								ServiceName: settings.IstioGateway,
								ServicePort: intstr.FromInt(80),
							},
						},
					},
				},
			},
		},
	}
	if tls {
		ingress.Spec.TLS = []v1beta1.IngressTLS{
			{
				Hosts:      []string{wildcardsDomain},
				SecretName: issuers.RioWildcardCerts,
			},
		}
	}

	os.Add(ingress)
	return nil
}
