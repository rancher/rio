package populate

import (
	"testing"

	"github.com/knative/pkg/apis/istio/v1alpha3"
	"github.com/rancher/rio/modules/test"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
	v1alpha32 "istio.io/api/networking/v1alpha3"
)

func TestServiceEntryFQDN(t *testing.T) {
	os := objectset.NewObjectSet()
	input := riov1.NewExternalService("default", "test", riov1.ExternalService{
		Spec: riov1.ExternalServiceSpec{
			FQDN: "www.foo.com",
		},
	})

	expected := constructors.NewServiceEntry(input.Namespace, input.Name, v1alpha3.ServiceEntry{
		Spec: v1alpha3.ServiceEntrySpec{
			Hosts:      []string{"www.foo.com"},
			Location:   int32(v1alpha32.ServiceEntry_MESH_EXTERNAL),
			Resolution: int32(v1alpha32.ServiceEntry_DNS),
			Ports: []v1alpha3.Port{
				{
					Protocol: v1alpha3.ProtocolHTTP,
					Number:   80,
					Name:     "http-80",
				},
			},
			Endpoints: []v1alpha3.ServiceEntry_Endpoint{
				{
					Address: "www.foo.com",
					Ports: map[string]uint32{
						"http": 80,
					},
				},
			},
		},
	})

	if err := ServiceEntry(input, os); err != nil {
		t.Fatal(err)
	}

	test.AssertObjects(t, expected, os)
}

func TestServiceEntryFQDNHttps(t *testing.T) {
	os := objectset.NewObjectSet()
	input := riov1.NewExternalService("default", "test", riov1.ExternalService{
		Spec: riov1.ExternalServiceSpec{
			FQDN: "https://www.foo.com",
		},
	})

	expected := constructors.NewServiceEntry(input.Namespace, input.Name, v1alpha3.ServiceEntry{
		Spec: v1alpha3.ServiceEntrySpec{
			Hosts:      []string{"www.foo.com"},
			Location:   int32(v1alpha32.ServiceEntry_MESH_EXTERNAL),
			Resolution: int32(v1alpha32.ServiceEntry_DNS),
			Ports: []v1alpha3.Port{
				{
					Protocol: v1alpha3.ProtocolHTTPS,
					Number:   443,
					Name:     "https-443",
				},
			},
			Endpoints: []v1alpha3.ServiceEntry_Endpoint{
				{
					Address: "www.foo.com",
					Ports: map[string]uint32{
						"https": 443,
					},
				},
			},
		},
	})

	if err := ServiceEntry(input, os); err != nil {
		t.Fatal(err)
	}

	test.AssertObjects(t, expected, os)
}

func TestServiceEntryFQDNCustomPort(t *testing.T) {
	os := objectset.NewObjectSet()
	input := riov1.NewExternalService("default", "test", riov1.ExternalService{
		Spec: riov1.ExternalServiceSpec{
			FQDN: "http://www.foo.com:8888",
		},
	})

	expected := constructors.NewServiceEntry(input.Namespace, input.Name, v1alpha3.ServiceEntry{
		Spec: v1alpha3.ServiceEntrySpec{
			Hosts:      []string{"www.foo.com"},
			Location:   int32(v1alpha32.ServiceEntry_MESH_EXTERNAL),
			Resolution: int32(v1alpha32.ServiceEntry_DNS),
			Ports: []v1alpha3.Port{
				{
					Protocol: v1alpha3.ProtocolHTTP,
					Number:   8888,
					Name:     "http-8888",
				},
			},
			Endpoints: []v1alpha3.ServiceEntry_Endpoint{
				{
					Address: "www.foo.com",
					Ports: map[string]uint32{
						"http": 8888,
					},
				},
			},
		},
	})

	if err := ServiceEntry(input, os); err != nil {
		t.Fatal(err)
	}

	test.AssertObjects(t, expected, os)
}
