package populate

import (
	"testing"

	"github.com/rancher/rio/modules/test"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestRouterIngress(t *testing.T) {
	os := objectset.NewObjectSet()

	systemNs := "rio-system-fake"
	domain := "foo.on-rio.io"
	certName := "rio-wildcard"

	input := riov1.NewRouter("default", "test", riov1.Router{
		Spec: riov1.RouterSpec{
			Routes: []riov1.RouteSpec{},
		},
	})

	expected := constructors.NewIngress(systemNs, "test-default", networkingv1beta1.Ingress{
		Spec: networkingv1beta1.IngressSpec{
			Rules: []networkingv1beta1.IngressRule{
				{
					Host: "test-default.foo.on-rio.io",
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
					Hosts:      []string{"*.foo.on-rio.io"},
					SecretName: certName,
				},
			},
		},
	})

	Ingress(systemNs, domain, certName, input, os)

	test.AssertObjects(t, expected, os)
}
