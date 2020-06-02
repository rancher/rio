package vsfactory

import (
	"fmt"

	"github.com/rancher/rio/pkg/constants"
	gatewayv1 "github.com/solo-io/gloo/projects/gateway/pkg/api/v1"
	solov1 "github.com/solo-io/gloo/projects/gateway/pkg/api/v1/kube/apis/gateway.solo.io/v1"
	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/gloo/projects/gloo/pkg/api/v1/core/matchers"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"istio.io/api/networking/v1alpha3"
	istiov1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
)

func (f *VirtualServiceFactory) InjectACME(vs *solov1.VirtualService) {
	for _, domain := range vs.Spec.VirtualHost.Domains {
		if _, err := f.publicDomainCache.Get(domain); err == nil {
			vs.Spec.VirtualHost.Routes = append([]*gatewayv1.Route{
				{
					Matchers: []*matchers.Matcher{
						{
							PathSpecifier: &matchers.Matcher_Prefix{
								Prefix: "/.well-known/acme-challenge",
							},
						},
					},
					Action: &gatewayv1.Route_RouteAction{
						RouteAction: &v1.RouteAction{
							Destination: &v1.RouteAction_Single{
								Single: &v1.Destination{
									DestinationType: &v1.Destination_Kube{
										Kube: &v1.KubernetesServiceDestination{
											Ref: core.ResourceRef{
												Name:      constants.AcmeSolverServicName,
												Namespace: f.systemNamespace,
											},
											Port: 8080,
										},
									},
								},
							},
						},
					},
				},
			}, vs.Spec.VirtualHost.Routes...)
		}
	}
	return
}

func (f *VirtualServiceFactory) InjectACMEIstio(vs *istiov1alpha3.VirtualService) {
	for _, domain := range vs.Spec.Hosts {
		if _, err := f.publicDomainCache.Get(domain); err == nil {
			vs.Spec.Http = append([]*v1alpha3.HTTPRoute{
				{
					Match: []*v1alpha3.HTTPMatchRequest{
						{
							Uri: &v1alpha3.StringMatch{
								MatchType: &v1alpha3.StringMatch_Prefix{
									Prefix: "/.well-known/acme-challenge",
								},
							},
						},
					},
					Route: []*v1alpha3.HTTPRouteDestination{
						{
							Destination: &v1alpha3.Destination{
								Host: fmt.Sprintf("%s.%s.svc.cluster.local", constants.AcmeSolverServicName, f.systemNamespace),
							},
						},
					},
				},
			}, vs.Spec.Http...)
		}
	}
}
