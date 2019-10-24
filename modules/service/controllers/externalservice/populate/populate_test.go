package populate

import (
	"testing"

	"github.com/rancher/rio/modules/test"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestServiceForExternalServiceFQDN(t *testing.T) {
	os := objectset.NewObjectSet()

	input := riov1.NewExternalService("default", "external", riov1.ExternalService{
		Spec: riov1.ExternalServiceSpec{
			FQDN: "www.foo.bar",
		},
	})

	expected := constructors.NewService("default", "external", v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"rio.cattle.io/service": "external",
			},
		},
		Spec: v1.ServiceSpec{
			Type:         v1.ServiceTypeExternalName,
			ExternalName: "www.foo.bar",
		},
	})

	if err := ServiceForExternalService(input, os); err != nil {
		t.Fatal(err)
	}

	test.AssertObjects(t, expected, os)
}

func TestServiceForExternalServiceIP(t *testing.T) {
	t.Skip("Fix post 0.6 RC")
	os := objectset.NewObjectSet()

	input := riov1.NewExternalService("default", "external", riov1.ExternalService{
		Spec: riov1.ExternalServiceSpec{
			IPAddresses: []string{
				"1.1.1.1:53",
				"2.2.2.2",
			},
		},
	})

	expectedService := constructors.NewService("default", "external", v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"rio.cattle.io/service": "external",
			},
		},
		Spec: v1.ServiceSpec{
			Type:      v1.ServiceTypeClusterIP,
			ClusterIP: v1.ClusterIPNone,
			Ports: []v1.ServicePort{
				{
					Name:     "http-53-53",
					Protocol: v1.ProtocolTCP,
					Port:     53,
				},
				{
					Name:     "http-80-80",
					Protocol: v1.ProtocolTCP,
					Port:     80,
				},
			},
		},
	})

	endpoint := constructors.NewEndpoints("default", "external", v1.Endpoints{
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP: "1.1.1.1",
					},
				},
				Ports: []v1.EndpointPort{
					{
						Name:     "http-53-53",
						Protocol: v1.ProtocolTCP,
						Port:     53,
					},
				},
			},
			{
				Addresses: []v1.EndpointAddress{
					{
						IP: "2.2.2.2",
					},
				},
				Ports: []v1.EndpointPort{
					{
						Name:     "http-80-80",
						Protocol: v1.ProtocolTCP,
						Port:     80,
					},
				},
			},
		},
	})

	if err := ServiceForExternalService(input, os); err != nil {
		t.Fatal(err)
	}

	test.AssertObjects(t, expectedService, os)

	test.AssertObjects(t, endpoint, os)
}
