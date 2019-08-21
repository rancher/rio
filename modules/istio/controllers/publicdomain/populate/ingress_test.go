package populate

import (
	"testing"

	"github.com/rancher/rio/modules/test"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestPublicDomainIngress(t *testing.T) {
	os := objectset.NewObjectSet()
	pd := adminv1.NewPublicDomain("default", "pd1", adminv1.PublicDomain{
		Spec: adminv1.PublicDomainSpec{
			SecretRef: v1.SecretReference{
				Name:      "pd1-secret",
				Namespace: "default",
			},
			DomainName: "www.foo.com",
		},
		Status: adminv1.PublicDomainStatus{
			IssuerName: "fake-issuer",
		},
	})

	systemNs := "rio-system-fake"
	ingressName := "pd1-41cf6"
	expected := constructors.NewIngress(systemNs, ingressName, networkingv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"certmanager.k8s.io/cluster-issuer": "fake-issuer",
			},
		},
		Spec: networkingv1beta1.IngressSpec{
			Rules: []networkingv1beta1.IngressRule{
				{
					Host: "www.foo.com",
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
					Hosts:      []string{"www.foo.com"},
					SecretName: "pd1-secret",
				},
			},
		},
	})

	Ingress(systemNs, pd, os)

	test.AssertObjects(t, expected, os)
}
