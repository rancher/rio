package populate

import (
	"testing"

	"github.com/knative/pkg/apis/istio/common/v1alpha1"
	"github.com/knative/pkg/apis/istio/v1alpha3"
	"github.com/rancher/rio/modules/test"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
)

func TestRouterForVirtualServices1(t *testing.T) {
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

	externalserviceMap := map[string]*riov1.ExternalService{
		"externalservice-foo-fqdn": riov1.NewExternalService("default", "externalservice-foo", riov1.ExternalService{
			Spec: riov1.ExternalServiceSpec{
				FQDN: "www.foo.bar",
			},
		}),
		"externalservice-foo-ip": riov1.NewExternalService("default", "externalservice-foo-ip", riov1.ExternalService{
			Spec: riov1.ExternalServiceSpec{
				IPAddresses: []string{
					"1.1.1.1:53",
					"1.2.3.4:8080",
					"1.2.3.5",
				},
			},
		}),
		"externalservice-foo-service": riov1.NewExternalService("default", "externalservice-foo", riov1.ExternalService{
			Spec: riov1.ExternalServiceSpec{
				Service: "anotherns/another",
			},
		}),
	}
	routerMap := map[string]*riov1.Router{}

	input := riov1.NewRouter("default", "test", riov1.Router{
		Spec: riov1.RouterSpec{
			Routes: []riov1.RouteSpec{
				{
					Matches: []riov1.Match{
						{
							Path: &riov1.StringMatch{
								Exact: "/foo",
							},
						},
						{
							Path: &riov1.StringMatch{
								Prefix: "/bar",
							},
						},
						{
							Path: &riov1.StringMatch{
								Regexp: "/[a-z]*",
							},
						},
						{
							Scheme: &riov1.StringMatch{
								Exact: "http",
							},
						},
						{
							Method: &riov1.StringMatch{
								Exact: "GET",
							},
						},
						{
							Headers: map[string]riov1.StringMatch{
								"FOO": {
									Exact: "BAR",
								},
								"FOO1": {
									Prefix: "BAR1",
								},
								"FOO2": {
									Regexp: "[a-z]*",
								},
							},
						},
						{
							Cookies: map[string]riov1.StringMatch{
								"USER": {
									Exact: "BAR",
								},
							},
						},
						{
							Port: &[]int{80}[0],
						},
					},
					To: []riov1.WeightedDestination{
						{
							Weight: 20,
							Destination: riov1.Destination{
								Service:   "service-foo",
								Namespace: "default",
								Revision:  "v0",
								Port:      &[]uint32{8080}[0],
							},
						},
						{
							Weight: 10,
							Destination: riov1.Destination{
								Service:   "externalservice-foo-fqdn",
								Namespace: "default",
								Port:      &[]uint32{8080}[0],
							},
						},
						{
							Weight: 10,
							Destination: riov1.Destination{
								Service:   "externalservice-foo-ip",
								Namespace: "default",
								Port:      &[]uint32{8080}[0],
							},
						},
						{
							Weight: 10,
							Destination: riov1.Destination{
								Service:   "externalservice-foo-service",
								Namespace: "default",
								Port:      &[]uint32{8080}[0],
							},
						},
					},
					Redirect: &riov1.Redirect{
						Host: "redirect-foo",
						Path: "/redirect-path",
					},
					Rewrite: &riov1.Rewrite{
						Host: "rewrite-host",
						Path: "/rewrite-host",
					},
					Headers: &v1alpha3.HeaderOperations{
						Add: map[string]string{
							"FOO": "BAR",
						},
						Remove: []string{
							"FOO1",
						},
						Set: map[string]string{
							"FOO2": "BAR2",
						},
					},
					RouteTraffic: riov1.RouteTraffic{
						Fault: &riov1.Fault{
							Percentage:  80,
							DelayMillis: 1000,
							Abort: riov1.Abort{
								HTTPStatus: 500,
							},
						},
						Mirror: &riov1.Destination{
							Service:   "mirror",
							Namespace: "default",
							Revision:  "v0",
							Port:      &[]uint32{8080}[0],
						},
						TimeoutMillis: &[]int{100}[0],
						Retry: &riov1.Retry{
							Attempts:      5,
							TimeoutMillis: 100,
						},
					},
				},
			},
		},
	})

	expected := constructors.NewVirtualService("default", "test", v1alpha3.VirtualService{
		Spec: v1alpha3.VirtualServiceSpec{
			Hosts: []string{
				"test",
				"test-default.foo.on-rio.io",
			},
			Gateways: []string{
				"mesh",
				"rio-gateway.rio-system-fake.svc.cluster.local",
			},
			HTTP: []v1alpha3.HTTPRoute{
				{
					Match: []v1alpha3.HTTPMatchRequest{
						{
							URI: &v1alpha1.StringMatch{
								Exact: "/foo",
							},
						},
						{
							URI: &v1alpha1.StringMatch{
								Prefix: "/bar",
							},
						},
						{
							URI: &v1alpha1.StringMatch{
								Regex: "/[a-z]*",
							},
						},
						{
							Scheme: &v1alpha1.StringMatch{
								Exact: "http",
							},
						},
						{
							Method: &v1alpha1.StringMatch{
								Exact: "GET",
							},
						},
						{
							Headers: map[string]v1alpha1.StringMatch{
								"FOO": {
									Exact: "BAR",
								},
								"FOO1": {
									Prefix: "BAR1",
								},
								"FOO2": {
									Regex: "[a-z]*",
								},
							},
						},
						{
							Headers: map[string]v1alpha1.StringMatch{
								"Cookie": {
									Exact: "USER=BAR",
								},
							},
						},
						{
							Port: 80,
						},
					},
					Timeout: "100ms",
					Rewrite: &v1alpha3.HTTPRewrite{
						URI:       "/rewrite-host",
						Authority: "rewrite-host",
					},
					Redirect: &v1alpha3.HTTPRedirect{
						URI:       "/redirect-path",
						Authority: "redirect-foo",
					},
					Headers: &v1alpha3.Headers{
						Request: &v1alpha3.HeaderOperations{
							Add: map[string]string{
								"FOO": "BAR",
							},
							Remove: []string{
								"FOO1",
								"l5d-remote-ip",
								"l5d-server-id",
							},
							Set: map[string]string{
								"FOO2":             "BAR2",
								"l5d-dst-override": "externalservice-foo-service.default.svc.cluster.local:8080",
							},
						},
					},
					Retries: &v1alpha3.HTTPRetry{
						Attempts:      5,
						PerTryTimeout: "100ms",
					},
					Fault: &v1alpha3.HTTPFaultInjection{
						Delay: &v1alpha3.InjectDelay{
							Percent:    80,
							FixedDelay: "1s",
						},
						Abort: &v1alpha3.InjectAbort{
							Percent:    80,
							HTTPStatus: 500,
						},
					},
					Mirror: &v1alpha3.Destination{
						Host:   "mirror.default.svc.cluster.local",
						Subset: "v0",
						Port: v1alpha3.PortSelector{
							Number: 8080,
						},
					},
					Route: []v1alpha3.HTTPRouteDestination{
						{
							Weight: 20,
							Destination: v1alpha3.Destination{
								Host:   "service-foo.default.svc.cluster.local",
								Subset: "v0",
								Port: v1alpha3.PortSelector{
									Number: 8080,
								},
							},
						},
						{
							Weight: 10,
							Destination: v1alpha3.Destination{
								Host: "www.foo.bar",
								Port: v1alpha3.PortSelector{
									Number: 8080,
								},
							},
						},
						{
							Weight: 10,
							Destination: v1alpha3.Destination{
								Host: "externalservice-foo-ip.default.svc.cluster.local",
								Port: v1alpha3.PortSelector{
									Number: 8080,
								},
							},
						},
						{
							Weight: 10,
							Destination: v1alpha3.Destination{
								Host: "another.anotherns.svc.cluster.local",
								Port: v1alpha3.PortSelector{
									Number: 8080,
								},
							},
						},
					},
				},
			},
		},
	})

	if err := VirtualServices(systemNs, clusterDomain, input, externalserviceMap, routerMap, os); err != nil {
		t.Fatal(err)
	}

	test.AssertObjects(t, expected, os)
}
