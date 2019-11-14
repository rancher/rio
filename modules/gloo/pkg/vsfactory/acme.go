package vsfactory

import (
	"strings"

	"github.com/solo-io/gloo/projects/gloo/pkg/api/v1/core/matchers"

	"github.com/sirupsen/logrus"
	gatewayv1 "github.com/solo-io/gloo/projects/gateway/pkg/api/v1"
	solov1 "github.com/solo-io/gloo/projects/gateway/pkg/api/v1/kube/apis/gateway.solo.io/v1"
	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"k8s.io/apimachinery/pkg/labels"
)

func (f *VirtualServiceFactory) InjectACME(vs *solov1.VirtualService) error {
	for _, domain := range vs.Spec.VirtualHost.Domains {
		ingresses, err := f.ingresses.List(f.systemNamespace, labels.Everything())
		if err != nil {
			return err
		}
		for _, ing := range ingresses {
			if len(ing.OwnerReferences) > 0 && strings.HasPrefix(ing.OwnerReferences[0].Name, domain) && len(ing.Spec.Rules) > 0 {
				if len(ing.Spec.Rules) > 0 && ing.Spec.Rules[0].HTTP != nil && len(ing.Spec.Rules[0].HTTP.Paths) > 0 {
					logrus.Infof("injecting acme http-01 path for domain %s", domain)
					vs.Spec.VirtualHost.Routes = append([]*gatewayv1.Route{
						{
							Matchers: []*matchers.Matcher{
								{
									PathSpecifier: &matchers.Matcher_Exact{
										Exact: ing.Spec.Rules[0].HTTP.Paths[0].Path,
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
														Name:      ing.Spec.Rules[0].HTTP.Paths[0].Backend.ServiceName,
														Namespace: ing.Namespace,
													},
													Port: uint32(ing.Spec.Rules[0].HTTP.Paths[0].Backend.ServicePort.IntVal),
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
		}
	}
	return nil
}
