package vsfactory

import (
	"net/url"
	"time"

	"github.com/gogo/protobuf/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	name2 "github.com/rancher/wrangler/pkg/name"
	gatewayv1 "github.com/solo-io/gloo/projects/gateway/pkg/api/v1"
	solov1 "github.com/solo-io/gloo/projects/gateway/pkg/api/v1/kube/apis/gateway.solo.io/v1"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/gloo/projects/gloo/pkg/api/v1/plugins/faultinjection"
	"github.com/solo-io/gloo/projects/gloo/pkg/api/v1/plugins/headers"
	"github.com/solo-io/gloo/projects/gloo/pkg/api/v1/plugins/hostrewrite"
	"github.com/solo-io/gloo/projects/gloo/pkg/api/v1/plugins/transformation"
	solovcorev1 "github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *VirtualServiceFactory) ForRouter(router *riov1.Router) ([]*solov1.VirtualService, error) {
	vs, err := vsForRouter(router)
	if err != nil {
		return nil, err
	}

	if err := f.InjectACME(vs); err != nil {
		return nil, err
	}

	result := []*solov1.VirtualService{
		vs,
	}

	tls, err := f.findTLS(router.Namespace, router.Name, "", vs.Spec.VirtualHost.Domains)
	if err != nil {
		return nil, err
	}

	for hostname, tls := range tls {
		result = append(result, tlsCopy(hostname, f.systemNamespace, tls, vs))
	}

	return result, nil
}

func vsForRouter(router *riov1.Router) (*solov1.VirtualService, error) {
	vs := &solov1.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: router.Namespace,
			Name:      router.Name,
		},
		Spec: gatewayv1.VirtualService{
			VirtualHost: &gatewayv1.VirtualHost{},
		},
	}

	domains, err := domains(router)
	if err != nil {
		return nil, err
	}

	vs.Spec.VirtualHost.Domains = domains

	for _, route := range router.Spec.Routes {
		vs.Spec.VirtualHost.Routes = append(vs.Spec.VirtualHost.Routes, routeToRoute(router.Namespace, route))
	}

	return vs, nil
}

func domains(router *riov1.Router) (result []string, err error) {
	seen := map[string]bool{}
	for _, endpoint := range router.Status.Endpoints {
		u, err := url.Parse(endpoint)
		if err != nil {
			return nil, err
		}

		if seen[u.Host] {
			continue
		}
		seen[u.Host] = true

		result = append(result, u.Host)
	}

	return
}

func headersToHeaders(hs *riov1.HeaderOperations) *headers.HeaderManipulation {
	result := &headers.HeaderManipulation{}

	for _, add := range hs.Add {
		header := &headers.HeaderValueOption{
			Header: &headers.HeaderValue{
				Key:   add.Name,
				Value: add.Value,
			},
			Append: &types.BoolValue{
				Value: true,
			},
		}
		result.RequestHeadersToAdd = append(result.RequestHeadersToAdd, header)
	}

	for _, add := range hs.Set {
		header := &headers.HeaderValueOption{
			Header: &headers.HeaderValue{
				Key:   add.Name,
				Value: add.Value,
			},
		}
		result.RequestHeadersToAdd = append(result.RequestHeadersToAdd, header)
	}

	result.RequestHeadersToRemove = hs.Remove
	return result
}

func routeToRoute(namespace string, route riov1.RouteSpec) (result *gatewayv1.Route) {
	result = &gatewayv1.Route{
		RoutePlugins: &gloov1.RoutePlugins{},
		Matcher:      matchToMatch(route.Match),
	}

	if route.Headers != nil {
		result.RoutePlugins = &gloov1.RoutePlugins{
			HeaderManipulation: headersToHeaders(route.Headers),
		}
	}

	if route.Fault != nil {
		addFault(route, result)
	} else if route.Rewrite != nil {
		addRewrite(route, result)
	} else if route.Redirect != nil {
		addRedirect(route, result)
	} else if route.Mirror != nil {
		addMirror(namespace, route, result)
	} else if len(route.To) == 1 {
		addSingleDestination(namespace, route, result)
	} else if len(route.To) > 1 {
		addMultiDestination(namespace, route, result)
	}

	return
}

func addMultiDestination(namespace string, route riov1.RouteSpec, gatewayRoute *gatewayv1.Route) {
	md := &gloov1.MultiDestination{}

	for _, to := range route.To {
		port := to.Port
		if port == 0 {
			port = 80
		}
		md.Destinations = append(md.Destinations, &gloov1.WeightedDestination{
			Destination: &gloov1.Destination{
				DestinationType: &gloov1.Destination_Kube{
					Kube: &gloov1.KubernetesServiceDestination{
						Ref:  *destinationToRef(namespace, &to.Destination),
						Port: port,
					},
				},
			},
			Weight: uint32(to.Weight),
		})
	}

	gatewayRoute.Action = &gatewayv1.Route_RouteAction{
		RouteAction: &gloov1.RouteAction{
			Destination: &gloov1.RouteAction_Multi{
				Multi: md,
			},
		},
	}
}

func addSingleDestination(namespace string, route riov1.RouteSpec, gatewayRoute *gatewayv1.Route) {
	port := route.To[0].Port
	if port == 0 {
		port = 80
	}
	gatewayRoute.Action = &gatewayv1.Route_RouteAction{
		RouteAction: &gloov1.RouteAction{
			Destination: &gloov1.RouteAction_Single{
				Single: &gloov1.Destination{
					DestinationType: &gloov1.Destination_Kube{
						Kube: &gloov1.KubernetesServiceDestination{
							Ref:  *destinationToRef(namespace, &route.To[0].Destination),
							Port: port,
						},
					},
				},
			},
		},
	}
}

func addMirror(namespace string, route riov1.RouteSpec, gatewayRoute *gatewayv1.Route) {
	gatewayRoute.Action = &gatewayv1.Route_DelegateAction{
		DelegateAction: destinationToRef(namespace, route.Mirror),
	}
}

func addRedirect(route riov1.RouteSpec, gatewayRoute *gatewayv1.Route) {
	ra := &gatewayv1.Route_RedirectAction{
		RedirectAction: &gloov1.RedirectAction{
			HostRedirect:  route.Redirect.Host,
			HttpsRedirect: route.Redirect.ToHTTPS,
		},
	}
	if route.Redirect.Path != "" {
		ra.RedirectAction.PathRewriteSpecifier = &gloov1.RedirectAction_PathRedirect{
			PathRedirect: route.Redirect.Path,
		}
	} else if route.Redirect.Prefix != "" {
		ra.RedirectAction.PathRewriteSpecifier = &gloov1.RedirectAction_PathRedirect{
			PathRedirect: route.Redirect.Path,
		}
	}
	gatewayRoute.Action = ra
}

func addRewrite(route riov1.RouteSpec, gatewayRoute *gatewayv1.Route) {
	if route.Rewrite.Path != "" {
		gatewayRoute.RoutePlugins.PrefixRewrite = &transformation.PrefixRewrite{
			PrefixRewrite: route.Rewrite.Path,
		}
	}

	if route.Rewrite.Host != "" {
		gatewayRoute.RoutePlugins.HostRewrite = &hostrewrite.HostRewrite{
			HostRewriteType: &hostrewrite.HostRewrite_HostRewrite{
				HostRewrite: route.Rewrite.Host,
			},
		}
	}
}

func addFault(route riov1.RouteSpec, gatewayRoute *gatewayv1.Route) {
	gatewayRoute.RoutePlugins.Faults = faultToFault(route.Fault)
}

func destinationToRef(namespace string, dest *riov1.Destination) *solovcorev1.ResourceRef {
	name := dest.App
	if dest.Version != "" {
		name = name2.SafeConcatName(name + "-" + dest.Version)
	}
	return &solovcorev1.ResourceRef{
		Name:      name,
		Namespace: namespace,
	}
}

func faultToFault(fault *riov1.Fault) *faultinjection.RouteFaults {
	if fault == nil {
		return nil
	}

	result := &faultinjection.RouteFaults{}

	if fault.AbortHTTPStatus > 0 {
		result.Abort = &faultinjection.RouteAbort{
			Percentage: float32(fault.Percentage),
			HttpStatus: uint32(fault.AbortHTTPStatus),
		}
	}

	if fault.DelayMillis > 0 {
		delay := time.Duration(fault.DelayMillis) * time.Millisecond
		result.Delay = &faultinjection.RouteDelay{
			Percentage: float32(fault.Percentage),
			FixedDelay: &delay,
		}
	}

	return result
}

func matchToMatch(match riov1.Match) (result *gloov1.Matcher) {
	result = &gloov1.Matcher{}

	if match.Path != nil {
		switch {
		case match.Path.Regexp != "":
			result.PathSpecifier = &gloov1.Matcher_Regex{
				Regex: match.Path.Regexp,
			}
		case match.Path.Prefix != "":
			result.PathSpecifier = &gloov1.Matcher_Prefix{
				Prefix: match.Path.Prefix,
			}
		case match.Path.Exact != "":
			result.PathSpecifier = &gloov1.Matcher_Exact{
				Exact: match.Path.Exact,
			}
		}
	} else {
		result.PathSpecifier = &gloov1.Matcher_Prefix{
			Prefix: "/",
		}
	}

	for _, match := range match.Headers {
		m := &gloov1.HeaderMatcher{
			Name: match.Name,
		}

		if match.Value != nil {
			switch {
			case match.Value.Regexp != "":
				m.Regex = true
				m.Value = match.Value.Regexp
			case match.Value.Prefix != "":
				m.Regex = true
				m.Value = "^" + match.Value.Prefix + ".*"
			case match.Value.Exact != "":
				m.Value = match.Value.Exact
			}
		}

		result.Headers = append(result.Headers, m)
	}

	result.Methods = match.Methods
	return
}
