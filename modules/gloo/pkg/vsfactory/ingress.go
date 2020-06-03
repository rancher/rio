package vsfactory

import (
	gloov1 "github.com/solo-io/gloo/projects/gateway/pkg/api/v1"
	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/gloo/projects/gloo/pkg/api/v1/core/matchers"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (f *VirtualServiceFactory) ForIngress(ingress *v1beta1.Ingress) ([]runtime.Object, error) {
	vs := newGlooVirtualService(ingress.Namespace, ingress.Name, nil)
	vs.Spec.VirtualHost.Routes = nil
	for _, rule := range ingress.Spec.Rules {
		vs.Spec.VirtualHost.Domains = append(vs.Spec.VirtualHost.Domains, rule.Host)
		for _, path := range rule.HTTP.Paths {
			vs.Spec.VirtualHost.Routes = append(vs.Spec.VirtualHost.Routes, &gloov1.Route{
				Matchers: []*matchers.Matcher{
					{
						PathSpecifier: &matchers.Matcher_Exact{
							Exact: path.Path,
						},
					},
				},
				Action: &gloov1.Route_RouteAction{
					RouteAction: &v1.RouteAction{
						Destination: &v1.RouteAction_Multi{
							Multi: &v1.MultiDestination{
								Destinations: []*v1.WeightedDestination{
									{
										Destination: destination(target{
											Name:      path.Backend.ServiceName,
											Namespace: ingress.Namespace,
											Port:      path.Backend.ServicePort.IntVal,
										}),
									},
								},
							},
						},
					},
				},
			})
		}
	}
	return []runtime.Object{vs}, nil
}
