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

func TestIngressWithVersion(t *testing.T) {
	os := objectset.NewObjectSet()

	systemNs := "rio-system-fake"
	domain := "foo.on-rio.io"
	certName := "rio-wildcard-fake"

	input := riov1.NewService("default", "test", riov1.Service{
		Spec: riov1.ServiceSpec{
			ServiceRevision: riov1.ServiceRevision{
				App:     "foo",
				Version: "v0",
			},
			PodConfig: riov1.PodConfig{
				Container: riov1.Container{
					Ports: []riov1.ContainerPort{
						{
							TargetPort: 80,
						},
					},
				},
			},
		},
	})

	expected := constructors.NewIngress(systemNs, "foo-v0-default", networkingv1beta1.Ingress{
		Spec: networkingv1beta1.IngressSpec{
			Rules: []networkingv1beta1.IngressRule{
				{
					Host: "foo-v0-default.foo.on-rio.io",
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

	Ingress(systemNs, domain, certName, false, input, os)

	test.AssertObjects(t, expected, os)
}

func TestIngressWithoutVersion(t *testing.T) {
	os := objectset.NewObjectSet()

	systemNs := "rio-system-fake"
	domain := "foo.on-rio.io"
	certName := "rio-wildcard-fake"

	input := riov1.NewService("default", "test", riov1.Service{
		Spec: riov1.ServiceSpec{
			ServiceRevision: riov1.ServiceRevision{
				App:     "foo",
				Version: "v0",
			},
			PodConfig: riov1.PodConfig{
				Container: riov1.Container{
					Ports: []riov1.ContainerPort{
						{
							TargetPort: 80,
						},
					},
				},
			},
		},
	})

	expected := constructors.NewIngress(systemNs, "foo-default", networkingv1beta1.Ingress{
		Spec: networkingv1beta1.IngressSpec{
			Rules: []networkingv1beta1.IngressRule{
				{
					Host: "foo-default.foo.on-rio.io",
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

	Ingress(systemNs, domain, certName, true, input, os)

	test.AssertObjects(t, expected, os)
}
