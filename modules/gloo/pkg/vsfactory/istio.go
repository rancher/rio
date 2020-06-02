package vsfactory

import (
	"fmt"

	"github.com/gogo/protobuf/types"
	"github.com/rancher/rio/modules/istio/controller/pkg"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	networkingv1alpha3 "istio.io/api/networking/v1alpha3"
	istiov1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newIstioVirtualService(namespace, name string, hosts []string, targets ...target) *istiov1alpha3.VirtualService {
	gateways := []string{
		fmt.Sprintf("%s/%s", constants.IstioSystemNamespace, constants.IstioRioGateway),
		"mesh",
	}

	var dests []*networkingv1alpha3.HTTPRouteDestination
	if len(targets) == 1 {
		targets[0].Weight = 100
	}
	for _, target := range targets {
		dests = append(dests, &networkingv1alpha3.HTTPRouteDestination{
			Destination: &networkingv1alpha3.Destination{
				Host:   fmt.Sprintf("%s.%s.svc.cluster.local", target.App, target.Namespace),
				Subset: target.Version,
			},
			Weight: int32(target.Weight),
		})
	}
	vs := &istiov1alpha3.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: networkingv1alpha3.VirtualService{
			Hosts:    append(hosts, name),
			Gateways: gateways,
			Http: []*networkingv1alpha3.HTTPRoute{
				{
					Match: []*networkingv1alpha3.HTTPMatchRequest{
						{
							Port: 80,
						},
						{
							Port: 443,
						},
					},
					Route: dests,
				},
			},
		},
	}

	return vs
}

func newIstioDestinationRule(namespace, name string, targets ...target) *istiov1alpha3.DestinationRule {
	subsets := []*networkingv1alpha3.Subset{
		{
			Name: "all",
			Labels: map[string]string{
				"app": name,
			},
		},
	}
	for _, target := range targets {
		subsets = append(subsets, &networkingv1alpha3.Subset{
			Name: target.Version,
			Labels: map[string]string{
				"app":     name,
				"version": target.Version,
			},
		})
	}
	return &istiov1alpha3.DestinationRule{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: networkingv1alpha3.DestinationRule{
			Host:    fmt.Sprintf("%s.%s.svc.cluster.local", name, namespace),
			Subsets: subsets,
		},
	}
}

func (f *VirtualServiceFactory) ForIstioRouter(router *riov1.Router) (*istiov1alpha3.VirtualService, error) {
	dms, err := pkg.Domains(router)
	if err != nil {
		return nil, err
	}

	vs := &istiov1alpha3.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      router.Name,
			Namespace: router.Namespace,
		},
		Spec: networkingv1alpha3.VirtualService{
			Gateways: []string{
				"mesh",
			},
			Hosts: dms,
		},
	}
	if !router.Spec.Internal {
		vs.Spec.Gateways = append(vs.Spec.Gateways,
			fmt.Sprintf("%s.%s.svc.cluster.local", constants.IstioRioGateway, constants.IstioSystemNamespace))
	}
	for _, route := range router.Spec.Routes {
		r := &networkingv1alpha3.HTTPRoute{}

		r.Match = []*networkingv1alpha3.HTTPMatchRequest{matchToIstioMatch(route.Match)}
		r.Redirect = redirectToIstio(route.Redirect)
		r.Rewrite = rewriteToIstio(route.Rewrite)
		r.Retries = retryToIstio(route.Retry)
		r.Fault = faultToIstio(route.Fault)
		r.Mirror = mirrorToIstio(route.Mirror)
		r.Headers = headerToIstio(route.Headers)
		r.Route = toIstioRoute(route.To)

		if route.TimeoutSeconds != nil {
			r.Timeout = &types.Duration{
				Seconds: int64(*route.TimeoutSeconds),
			}
		}
		vs.Spec.Http = []*networkingv1alpha3.HTTPRoute{r}
	}
	f.InjectACMEIstio(vs)

	return vs, nil
}

func toIstioRoute(to []riov1.WeightedDestination) []*networkingv1alpha3.HTTPRouteDestination {
	var r []*networkingv1alpha3.HTTPRouteDestination
	for _, t := range to {
		r = append(r, &networkingv1alpha3.HTTPRouteDestination{
			Destination: &networkingv1alpha3.Destination{
				Host:   t.App,
				Subset: t.Version,
			},
			Weight: int32(t.Weight),
		})
	}
	return r
}

func headerToIstio(header *riov1.HeaderOperations) *networkingv1alpha3.Headers {
	if header == nil {
		return nil
	}

	return &networkingv1alpha3.Headers{
		Request: &networkingv1alpha3.Headers_HeaderOperations{
			Set:    nameValuesToMap(header.Set),
			Add:    nameValuesToMap(header.Add),
			Remove: header.Remove,
		},
	}
}

func mirrorToIstio(mirror *riov1.Destination) *networkingv1alpha3.Destination {
	if mirror == nil {
		return nil
	}

	return &networkingv1alpha3.Destination{
		Host:   mirror.App,
		Subset: mirror.Version,
		Port:   &networkingv1alpha3.PortSelector{Number: mirror.Port},
	}
}

func nameValuesToMap(headers []riov1.NameValue) map[string]string {
	r := map[string]string{}
	for _, header := range headers {
		r[header.Name] = header.Value
	}
	return r
}

func faultToIstio(fault *riov1.Fault) *networkingv1alpha3.HTTPFaultInjection {
	if fault == nil {
		return nil
	}

	f := &networkingv1alpha3.HTTPFaultInjection{
		Delay: &networkingv1alpha3.HTTPFaultInjection_Delay{
			HttpDelayType: &networkingv1alpha3.HTTPFaultInjection_Delay_FixedDelay{
				FixedDelay: &types.Duration{
					Nanos: int32(fault.DelayMillis * 100),
				}},
			Percentage: &networkingv1alpha3.Percent{
				Value: float64(fault.Percentage),
			},
		},
	}

	if fault.AbortHTTPStatus != 0 {
		f.Abort = &networkingv1alpha3.HTTPFaultInjection_Abort{
			Percentage: &networkingv1alpha3.Percent{
				Value: float64(fault.Percentage),
			},
			ErrorType: &networkingv1alpha3.HTTPFaultInjection_Abort_HttpStatus{
				HttpStatus: int32(fault.AbortHTTPStatus),
			},
		}
	}
	return f
}

func retryToIstio(retry *riov1.Retry) *networkingv1alpha3.HTTPRetry {
	if retry == nil {
		return nil
	}

	return &networkingv1alpha3.HTTPRetry{
		Attempts:      int32(retry.Attempts),
		PerTryTimeout: &types.Duration{Seconds: int64(retry.TimeoutSeconds)},
	}
}

func rewriteToIstio(rewrite *riov1.Rewrite) *networkingv1alpha3.HTTPRewrite {
	if rewrite == nil {
		return nil
	}
	return &networkingv1alpha3.HTTPRewrite{
		Uri:       rewrite.Path,
		Authority: rewrite.Host,
	}
}

func redirectToIstio(redirect *riov1.Redirect) *networkingv1alpha3.HTTPRedirect {
	if redirect == nil {
		return nil
	}
	return &networkingv1alpha3.HTTPRedirect{
		Authority: redirect.Host,
		Uri:       redirect.Path,
	}
}

func matchToIstioMatch(match riov1.Match) *networkingv1alpha3.HTTPMatchRequest {
	return &networkingv1alpha3.HTTPMatchRequest{
		Uri:     stringMatch(match.Path),
		Scheme:  stringMatch(match.Schema),
		Method:  methodToMatch(match.Methods),
		Headers: headerMatch(match.Headers),
	}
}

// only support one match
func methodToMatch(methods []string) *networkingv1alpha3.StringMatch {
	if len(methods) == 0 {
		return nil
	}
	return &networkingv1alpha3.StringMatch{
		MatchType: &networkingv1alpha3.StringMatch_Exact{Exact: methods[0]},
	}
}

func headerMatch(headers []riov1.HeaderMatch) map[string]*networkingv1alpha3.StringMatch {
	r := map[string]*networkingv1alpha3.StringMatch{}
	for _, hm := range headers {
		r[hm.Name] = stringMatch(hm.Value)
	}
	return r
}

func stringMatch(sm *riov1.StringMatch) *networkingv1alpha3.StringMatch {
	if sm == nil {
		return nil
	}
	if sm.Prefix != "" {
		return &networkingv1alpha3.StringMatch{
			MatchType: &networkingv1alpha3.StringMatch_Prefix{Prefix: sm.Prefix},
		}
	}

	if sm.Regexp != "" {
		return &networkingv1alpha3.StringMatch{
			MatchType: &networkingv1alpha3.StringMatch_Regex{Regex: sm.Regexp},
		}
	}

	if sm.Exact != "" {
		return &networkingv1alpha3.StringMatch{
			MatchType: &networkingv1alpha3.StringMatch_Exact{Exact: sm.Exact},
		}
	}

	return nil
}
