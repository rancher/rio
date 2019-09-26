package populate

import (
	"testing"

	"github.com/knative/pkg/apis/istio/v1alpha3"
	"github.com/rancher/rio/modules/test"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
)

func TestVirtualServices(t *testing.T) {
	os := objectset.NewObjectSet()

	systemNs := "rio-system-fake"
	clusterDomain := adminv1.NewClusterDomain(systemNs, "cluster-domain", adminv1.ClusterDomain{
		Spec: adminv1.ClusterDomainSpec{
			SecretRef: v1.SecretReference{
				Name:      "rio-wildcard-fake",
				Namespace: "default",
			},
		},
		Status: adminv1.ClusterDomainStatus{
			ClusterDomain: "foo.on-rio.io",
		},
	})

	input := riov1.NewService("default", "test", riov1.Service{
		Spec: riov1.ServiceSpec{
			Autoscale: &riov1.AutoscaleConfig{
				MinReplicas: &[]int{0}[0],
				MaxReplicas: &[]int{10}[0],
				Concurrency: &[]int{10}[0],
			},
			App:     "foo",
			Version: "v0",
			PodConfig: riov1.PodConfig{
				Container: riov1.Container{
					Image: "test",
					Ports: []riov1.ContainerPort{
						{
							Port:       80,
							TargetPort: 80,
						},
					},
				},
			},
		},
	})

	expect := constructors.NewVirtualService("default", "foo-v0", v1alpha3.VirtualService{
		Spec: v1alpha3.VirtualServiceSpec{
			Hosts: []string{
				"foo-v0-default.foo.on-rio.io",
			},
			Gateways: []string{
				"mesh",
				"rio-gateway.rio-system-fake.svc.cluster.local",
			},
			HTTP: []v1alpha3.HTTPRoute{
				{
					Headers: &v1alpha3.Headers{
						Request: &v1alpha3.HeaderOperations{
							Set:    map[string]string{"l5d-dst-override": "foo-v0.default.svc.cluster.local:80"},
							Remove: []string{"l5d-remote-ip", "l5d-server-id"},
						},
					},
					Match: []v1alpha3.HTTPMatchRequest{
						{
							Gateways: []string{
								"mesh",
								"rio-gateway.rio-system-fake.svc.cluster.local",
							},
							Port: 9080,
						},
						{
							Gateways: []string{
								"mesh",
								"rio-gateway.rio-system-fake.svc.cluster.local",
							},
							Port: 9443,
						},
						{
							Gateways: []string{
								"mesh",
							},
							Port: 80,
						},
					},
					Route: []v1alpha3.HTTPRouteDestination{
						{
							Destination: v1alpha3.Destination{
								Host: "foo.default.svc.cluster.local",
								Port: v1alpha3.PortSelector{
									Number: 80,
								},
								Subset: "v0",
							},
							Weight: 100,
						},
					},
				},
			},
		},
	})

	if err := VirtualServices(systemNs, clusterDomain, input, os); err != nil {
		t.Fatal(err)
	}

	test.AssertObjects(t, expect, os)
}
