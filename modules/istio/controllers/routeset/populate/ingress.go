package populate

import (
	"fmt"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Ingress(namespace, domain, certName string, route *riov1.Router, os *objectset.ObjectSet) {
	host := fmt.Sprintf("%s-%s.%s", route.Name, route.Namespace, domain)

	ingress := constructors.NewIngress(namespace, fmt.Sprintf("%s-%s", route.Name, route.Namespace), networkingv1beta1.Ingress{
		Spec: networkingv1beta1.IngressSpec{
			Rules: []networkingv1beta1.IngressRule{
				{
					Host: host,
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
					Hosts:      []string{fmt.Sprintf("*.%s", domain)},
					SecretName: certName,
				},
			},
		},
	})

	os.Add(ingress)
	return
}
