package populate

import (
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/name"
	"github.com/rancher/wrangler/pkg/objectset"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Ingress(systemNamespace string, pd *adminv1.PublicDomain, os *objectset.ObjectSet) {
	ingress := constructors.NewIngress(systemNamespace, name.SafeConcatName(pd.Name, name.Hex(pd.Spec.DomainName, 5)), networkingv1beta1.Ingress{
		Spec: networkingv1beta1.IngressSpec{
			Rules: []networkingv1beta1.IngressRule{
				{
					Host: pd.Spec.DomainName,
					IngressRuleValue: networkingv1beta1.IngressRuleValue{
						HTTP: &networkingv1beta1.HTTPIngressRuleValue{
							Paths: []networkingv1beta1.HTTPIngressPath{
								{
									Backend: networkingv1beta1.IngressBackend{
										ServiceName: constants.IstioGateway,
										ServicePort: intstr.FromInt(80),
									},
								},
							},
						},
					},
				},
			},
			TLS: []networkingv1beta1.IngressTLS{
				{
					Hosts:      []string{pd.Spec.DomainName},
					SecretName: pd.Spec.SecretRef.Name,
				},
			},
		},
	})
	ingress.Annotations = make(map[string]string)
	ingress.Annotations["certmanager.k8s.io/cluster-issuer"] = pd.Status.IssuerName

	os.Add(ingress)
	return
}
