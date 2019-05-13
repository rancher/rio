package populate

import (
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/pkg/name"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/wrangler/pkg/objectset"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Ingress(systemNamespace string, publicdomain *riov1.PublicDomain, os *objectset.ObjectSet) error {
	ingress := constructors.NewIngress(systemNamespace, name.SafeConcatName(publicdomain.Name, name.Hex(publicdomain.Spec.DomainName, 5)), v1beta1.Ingress{})

	ingress.Spec.Rules = []v1beta1.IngressRule{
		{
			Host: publicdomain.Spec.DomainName,
			IngressRuleValue: v1beta1.IngressRuleValue{
				HTTP: &v1beta1.HTTPIngressRuleValue{
					Paths: []v1beta1.HTTPIngressPath{
						{
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
	ingress.Spec.TLS = append(ingress.Spec.TLS, v1beta1.IngressTLS{
		Hosts:      []string{publicdomain.Spec.DomainName},
		SecretName: publicdomain.Spec.SecretRef.Name,
	})
	os.Add(ingress)
	return nil
}
