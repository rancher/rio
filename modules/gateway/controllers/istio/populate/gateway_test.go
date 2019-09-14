package populate

import (
	"testing"

	"github.com/knative/pkg/apis/istio/v1alpha3"
	"github.com/rancher/rio/modules/test"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
)

func TestGatewayWithoutPublicdomains(t *testing.T) {
	constants.InstallMode = constants.InstallModeSvclb
	os := objectset.NewObjectSet()

	systemNs := "rio-system-fake"
	expected := constructors.NewGateway(systemNs, constants.RioGateway, v1alpha3.Gateway{
		Spec: v1alpha3.GatewaySpec{
			Selector: map[string]string{
				"app": constants.GatewayName,
			},
			Servers: []v1alpha3.Server{
				{
					Port: v1alpha3.Port{
						Protocol: v1alpha3.ProtocolHTTP,
						Number:   80,
						Name:     "http-80",
					},
					Hosts: []string{"*"},
				},
				{
					Port: v1alpha3.Port{
						Protocol: v1alpha3.ProtocolHTTPS,
						Number:   443,
						Name:     "https-443",
					},
					Hosts: []string{"*"},
					TLS: &v1alpha3.TLSOptions{
						Mode:           v1alpha3.TLSModeSimple,
						CredentialName: "rio-certs",
					},
				},
			},
		},
	})

	Gateway(systemNs, "foo.on-rio.io", "rio-certs", nil, os)

	test.AssertObjects(t, expected, os)
}

func TestGatewayWithPublicdomains(t *testing.T) {
	constants.InstallMode = constants.InstallModeSvclb
	os := objectset.NewObjectSet()

	systemNs := "rio-system-fake"
	publicDomains := []*adminv1.PublicDomain{
		adminv1.NewPublicDomain("default", "pd1", adminv1.PublicDomain{
			Spec: adminv1.PublicDomainSpec{
				SecretRef: v1.SecretReference{
					Name:      "pd1-secret",
					Namespace: "default",
				},
				DomainName: "www.foo.com",
			},
		}),
		adminv1.NewPublicDomain("default", "pd2", adminv1.PublicDomain{
			Spec: adminv1.PublicDomainSpec{
				SecretRef: v1.SecretReference{
					Name:      "pd2-secret",
					Namespace: "default",
				},
				DomainName: "www.bar.com",
			},
		}),
	}

	expected := constructors.NewGateway(systemNs, constants.RioGateway, v1alpha3.Gateway{
		Spec: v1alpha3.GatewaySpec{
			Selector: map[string]string{
				"app": constants.GatewayName,
			},
			Servers: []v1alpha3.Server{
				{
					Port: v1alpha3.Port{
						Protocol: v1alpha3.ProtocolHTTP,
						Number:   80,
						Name:     "http-80",
					},
					Hosts: []string{"*"},
				},
				{
					Port: v1alpha3.Port{
						Protocol: v1alpha3.ProtocolHTTPS,
						Number:   443,
						Name:     "https-443",
					},
					Hosts: []string{"*"},
					TLS: &v1alpha3.TLSOptions{
						Mode:           v1alpha3.TLSModeSimple,
						CredentialName: "rio-certs",
					},
				},
				{
					Port: v1alpha3.Port{
						Protocol: v1alpha3.ProtocolHTTPS,
						Number:   443,
						Name:     "pd1-https-443",
					},
					Hosts: []string{"www.foo.com"},
					TLS: &v1alpha3.TLSOptions{
						Mode:           v1alpha3.TLSModeSimple,
						CredentialName: "pd1-secret",
					},
				},
				{
					Port: v1alpha3.Port{
						Protocol: v1alpha3.ProtocolHTTPS,
						Number:   443,
						Name:     "pd2-https-443",
					},
					Hosts: []string{"www.bar.com"},
					TLS: &v1alpha3.TLSOptions{
						Mode:           v1alpha3.TLSModeSimple,
						CredentialName: "pd2-secret",
					},
				},
			},
		},
	})

	Gateway(systemNs, "foo.on-rio.io", "rio-certs", publicDomains, os)

	test.AssertObjects(t, expected, os)
}
