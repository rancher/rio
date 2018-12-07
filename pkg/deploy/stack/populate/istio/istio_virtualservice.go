package istio

import (
	"fmt"
	"hash/adler32"
	"sort"
	"strconv"
	"strings"
	"time"

	google_protobuf "github.com/gogo/protobuf/types"
	service2 "github.com/rancher/rio/api/service"
	"github.com/rancher/rio/pkg/deploy/stack/input"
	"github.com/rancher/rio/pkg/deploy/stack/output"
	"github.com/rancher/rio/pkg/deploy/stack/populate/containerlist"
	"github.com/rancher/rio/pkg/deploy/stack/populate/service"
	"github.com/rancher/rio/pkg/deploy/stack/populate/servicelabels"
	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"istio.io/api/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	privateGw              = "mesh"
	PublicDomainAnnotation = "rio.cattle.io/publicDomain"
)

func virtualservices(stack *input.Stack) ([]*output.IstioObject, error) {
	var result []*output.IstioObject

	services, err := service.CollectionServices(stack.Services)
	if err != nil {
		return nil, err
	}

	for name, service := range services {
		svcs := vsFromService(stack, name, service)
		result = append(result, svcs...)
	}

	routesets := stack.RouteSet
	svcs := vsFromRoutesets(stack, routesets)
	result = append(result, svcs...)

	return result, nil
}

func vsRoutes(publicPorts map[string]bool, service *v1.Service, dests []dest) ([]*v1alpha3.HTTPRoute, bool) {
	external := false
	var result []*v1alpha3.HTTPRoute

	// add https challenge match
	pb := &v1.PortBinding{
		Port:       80,
		TargetPort: 8089,
		Protocol:   "http",
	}
	for _, publicDomain := range strings.Split(service.Annotations[PublicDomainAnnotation], ",") {
		if publicDomain == "" {
			continue
		}
		ds := []dest{
			{
				host:   fmt.Sprintf("%s.rio-system.svc.cluster.local", fmt.Sprintf("cm-acme-http-solver-%d", adler32.Checksum([]byte(publicDomain)))),
				subset: "latest",
				weight: 100,
			},
		}
		_, route := newRoute(GetPublicGateway(), true, pb, ds, false)
		route.Match[0].Uri = &v1alpha3.StringMatch{
			MatchType: &v1alpha3.StringMatch_Prefix{
				Prefix: "/.well-known/acme-challenge/",
			},
		}
		route.Match[0].Authority = &v1alpha3.StringMatch{
			MatchType: &v1alpha3.StringMatch_Prefix{
				Prefix: publicDomain,
			},
		}
		result = append(result, route)
	}

	containerlist.ForService(service)
	for _, con := range containerlist.ForService(service) {
		for _, exposed := range con.ExposedPorts {
			_, route := newRoute(GetPublicGateway(), false, &exposed.PortBinding, dests, true)
			if route != nil {
				result = append(result, route)
			}
		}

		for _, binding := range con.PortBindings {
			publicPort, route := newRoute(GetPublicGateway(), true, &binding, dests, true)
			if route != nil {
				external = true
				publicPorts[publicPort] = true
				result = append(result, route)
			}
		}
	}

	return result, external
}

func newRoute(externalGW string, published bool, portBinding *v1.PortBinding, dests []dest, appendHttps bool) (string, *v1alpha3.HTTPRoute) {
	if portBinding.Protocol != "http" && portBinding.Protocol != "https" {
		return "", nil
	}

	gw := []string{privateGw}
	if published {
		gw = append(gw, externalGW)
	}

	httpPort, _ := strconv.ParseUint(settings.DefaultHTTPOpenPort.Get(), 10, 64)
	httpsPort, _ := strconv.ParseUint(settings.DefaultHTTPSOpenPort.Get(), 10, 64)
	matches := []*v1alpha3.HTTPMatchRequest{
		{
			Port:     uint32(httpPort),
			Gateways: gw,
		},
	}
	if appendHttps {
		matches = append(matches,
			&v1alpha3.HTTPMatchRequest{
				Port:     uint32(httpsPort),
				Gateways: gw,
			})
	}
	route := &v1alpha3.HTTPRoute{
		Match: matches,
	}

	for _, dest := range dests {
		route.Route = append(route.Route, &v1alpha3.DestinationWeight{
			Destination: &v1alpha3.Destination{
				Host:   dest.host,
				Subset: dest.subset,
				Port: &v1alpha3.PortSelector{
					Port: &v1alpha3.PortSelector_Number{
						Number: uint32(portBinding.TargetPort),
					},
				},
			},
			Weight: dest.weight,
		})
	}

	sourcePort := httpPort
	if portBinding.Protocol == "https" {
		sourcePort = httpsPort
	}
	return fmt.Sprintf("%v/%s", sourcePort, portBinding.Protocol), route
}

type dest struct {
	host, subset string
	weight       int32
}

func destsForService(name string, service *output.ServiceSet) []dest {
	latestWeight := 100
	result := []dest{
		{
			host:   name,
			subset: service.Service.Spec.Revision.Version,
		},
	}

	for _, rev := range service.Revisions {
		if latestWeight == 0 {
			// no more weight left
			continue
		}

		weight := min(rev.Spec.Revision.Weight, 100)
		if weight <= 0 {
			continue
		}

		weight = min(weight, latestWeight)
		latestWeight -= weight

		result = append(result, dest{
			host:   rev.Name,
			weight: int32(weight),
			subset: rev.Spec.Revision.Version,
		})
	}

	result[0].weight = int32(latestWeight)
	if result[0].weight == 0 && len(result) > 1 {
		return result[1:]
	}
	return result
}

func min(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func vsFromService(stack *input.Stack, name string, service *output.ServiceSet) []*output.IstioObject {
	var result []*output.IstioObject

	serviceVS := vsFromSpec(stack, name, stack.Namespace, service.Service, destsForService(name, service)...)
	if serviceVS != nil {
		result = append(result, serviceVS)
	}

	for _, rev := range service.Revisions {
		revVs := vsFromSpec(stack, rev.Name, stack.Namespace, rev, dest{
			host:   rev.Name,
			subset: rev.Spec.Revision.Version,
			weight: 100,
		})
		if revVs != nil {
			result = append(result, revVs)
		}
	}

	return result
}

func vsFromSpec(stack *input.Stack, name, namespace string, service *v1.Service, dests ...dest) *output.IstioObject {
	publicPorts := map[string]bool{}

	routes, external := vsRoutes(publicPorts, service, dests)
	if len(routes) == 0 {
		return nil
	}

	vs := newVirtualService(stack, service)
	spec := &v1alpha3.VirtualService{
		Hosts:    []string{name},
		Gateways: []string{privateGw},
		Http:     routes,
	}
	vs.Spec = spec

	if external && len(publicPorts) > 0 {
		externalGW := GetPublicGateway()
		externalHost := getExternalDomain(name, namespace, stack.Project)
		spec.Gateways = append(spec.Gateways, externalGW)
		spec.Hosts = append(spec.Hosts, externalHost)

		var portList []string
		for p := range publicPorts {
			portList = append(portList, p)
		}
		sort.Slice(portList, func(i, j int) bool {
			return portList[i] < portList[j]
		})

		vs.Annotations["rio.cattle.io/ports"] = strings.Join(portList, ",")
	}

	if service.Annotations[PublicDomainAnnotation] != "" {
		spec.Hosts = append(spec.Hosts, strings.Split(service.Annotations[PublicDomainAnnotation], ",")...)
	}

	return vs
}

func vsFromRoutesets(stack *input.Stack, routesets []*v1.RouteSet) []*output.IstioObject {
	result := make([]*output.IstioObject, 0)
	externalServiceMap := map[string]*v1.ExternalService{}
	for _, e := range stack.ExternalServices {
		externalServiceMap[e.Name] = e
	}
	for _, routeset := range routesets {
		ns := namespace.StackNamespace(stack.Stack.Namespace, stack.Stack.Name)
		vs := newVirtualServiceGeneric(stack, routeset.Name, ns)
		spec := &v1alpha3.VirtualService{
			Gateways: []string{privateGw, GetPublicGateway()},
			Hosts:    []string{routeset.Name, getExternalDomain(routeset.Name, stack.Stack.Name, stack.Project)},
		}
		// populate http routing
		for _, routeSpec := range routeset.Spec.Routes {
			httpRoute := &v1alpha3.HTTPRoute{}
			// populate destinations
			for _, dest := range routeSpec.To {
				if svc, ok := externalServiceMap[dest.Service]; ok {
					httpRoute.Route = append(httpRoute.Route, destWeightForExternalService(dest, svc))
				} else {
					httpRoute.Route = append(httpRoute.Route, destWeightForService(dest))
				}
			}

			// populate matches
			for _, match := range routeSpec.Matches {
				httpMatch := &v1alpha3.HTTPMatchRequest{}
				httpMatch.Uri = populateStringMatch(match.Path)
				httpMatch.Scheme = populateStringMatch(match.Scheme)
				httpMatch.Method = populateStringMatch(match.Method)
				httpMatch.Port = uint32(match.Port)
				httpMatch.Headers = make(map[string]*v1alpha3.StringMatch, 0)
				for name, cookie := range match.Cookies {
					httpMatch.Headers[name] = populateStringMatch(cookie)
				}
				for name, value := range match.Headers {
					httpMatch.Headers[name] = populateStringMatch(value)
				}
				httpRoute.Match = append(httpRoute.Match, httpMatch)
			}
			if len(httpRoute.Match) == 0 {
				httpRoute.Match = []*v1alpha3.HTTPMatchRequest{
					{
						Gateways: []string{privateGw, GetPublicGateway()},
						Port:     80,
					},
					{
						Gateways: []string{privateGw, GetPublicGateway()},
						Port:     443,
					},
				}
			}

			if len(routeSpec.AddHeaders) > 0 {
				httpRoute.AppendHeaders = convertEnvFromSliceToMap(routeSpec.AddHeaders)
			}

			if routeSpec.Redirect != nil {
				httpRoute.Redirect = &v1alpha3.HTTPRedirect{
					Uri:       routeSpec.Redirect.Path,
					Authority: routeSpec.Redirect.Host,
				}
			}

			if routeSpec.Rewrite != nil {
				httpRoute.Rewrite = &v1alpha3.HTTPRewrite{
					Uri:       routeSpec.Rewrite.Path,
					Authority: routeSpec.Rewrite.Host,
				}
			}

			// fault handling
			if routeSpec.Fault != nil {
				httpRoute.Fault = &v1alpha3.HTTPFaultInjection{
					Delay: &v1alpha3.HTTPFaultInjection_Delay{
						Percent: int32(routeSpec.Fault.Percentage),
						HttpDelayType: &v1alpha3.HTTPFaultInjection_Delay_FixedDelay{
							FixedDelay: google_protobuf.DurationProto(time.Millisecond * time.Duration(routeSpec.Fault.DelayMillis)),
						},
					},
					Abort: populateHttpAbort(routeSpec.Fault),
				}
			}

			if routeSpec.TimeoutMillis != 0 {
				httpRoute.Timeout = google_protobuf.DurationProto(time.Millisecond * time.Duration(routeSpec.TimeoutMillis))
			}

			if routeSpec.Mirror != nil {
				httpRoute.Mirror = &v1alpha3.Destination{
					Host:   getExternalDomain(routeSpec.Mirror.Service, routeSpec.Mirror.Stack, stack.Project),
					Subset: routeSpec.Mirror.Revision,
					Port: &v1alpha3.PortSelector{
						Port: &v1alpha3.PortSelector_Number{
							Number: uint32(routeSpec.Mirror.Port),
						},
					},
				}
			}

			if routeSpec.Retry != nil {
				httpRoute.Retries = &v1alpha3.HTTPRetry{
					Attempts:      int32(routeSpec.Retry.Attempts),
					PerTryTimeout: google_protobuf.DurationProto(time.Millisecond * time.Duration(routeSpec.Retry.TimeoutMillis)),
				}
			}

			if routeset.Annotations[PublicDomainAnnotation] != "" {
				spec.Hosts = append(spec.Hosts, strings.Split(routeset.Annotations[PublicDomainAnnotation], ",")...)
			}

			spec.Http = append(spec.Http, httpRoute)
		}

		// set port to 80 for virtual services that are created from gateway
		vs.Annotations["rio.cattle.io/ports"] = "80"
		vs.Spec = spec
		result = append(result, vs)
	}
	return result
}

func populateHttpAbort(fault *v1.Fault) *v1alpha3.HTTPFaultInjection_Abort {
	abort := &v1alpha3.HTTPFaultInjection_Abort{
		Percent: int32(fault.Percentage),
	}
	if fault.Abort.GRPCStatus != "" {
		abort.ErrorType = &v1alpha3.HTTPFaultInjection_Abort_GrpcStatus{
			GrpcStatus: fault.Abort.GRPCStatus,
		}
	} else if fault.Abort.HTTP2Status != "" {
		abort.ErrorType = &v1alpha3.HTTPFaultInjection_Abort_Http2Error{
			Http2Error: fault.Abort.HTTP2Status,
		}
	} else if fault.Abort.HTTPStatus != 0 {
		abort.ErrorType = &v1alpha3.HTTPFaultInjection_Abort_HttpStatus{
			HttpStatus: int32(fault.Abort.HTTPStatus),
		}
	}
	return abort
}

func populateStringMatch(match v1.StringMatch) *v1alpha3.StringMatch {
	m := &v1alpha3.StringMatch{}
	if match.Exact != "" {
		m.MatchType = getExactMatch(match)
	} else if match.Prefix != "" {
		m.MatchType = getPrefixMatch(match)
	} else if match.Regexp != "" {
		m.MatchType = getRegexpMatch(match)
	}
	return m
}

func convertEnvFromSliceToMap(envs []string) map[string]string {
	m := map[string]string{}
	for _, env := range envs {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}
	return m
}

func getExactMatch(match v1.StringMatch) *v1alpha3.StringMatch_Exact {
	return &v1alpha3.StringMatch_Exact{
		Exact: match.Exact,
	}
}

func getPrefixMatch(match v1.StringMatch) *v1alpha3.StringMatch_Prefix {
	return &v1alpha3.StringMatch_Prefix{
		Prefix: match.Prefix,
	}
}

func getRegexpMatch(match v1.StringMatch) *v1alpha3.StringMatch_Regex {
	return &v1alpha3.StringMatch_Regex{
		Regex: match.Regexp,
	}
}

func GetPublicGateway() string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", settings.IstioGateway, settings.RioSystemNamespace)
}

func getExternalDomain(name, namespace, project string) string {
	return fmt.Sprintf("%s.%s", service2.HashIfNeed(name, strings.SplitN(namespace, "-", 2)[0], project), settings.ClusterDomain.Get())
}

func newVirtualService(stack *input.Stack, service *v1.Service) *output.IstioObject {
	return &output.IstioObject{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "VirtualService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        service.Name,
			Namespace:   service.Namespace,
			Annotations: map[string]string{},
			Labels:      servicelabels.RioOnlyServiceLabels(stack, service),
		},
	}
}

func newVirtualServiceGeneric(stack *input.Stack, name, namespace string) *output.IstioObject {
	return &output.IstioObject{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "VirtualService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: map[string]string{},
			Labels: map[string]string{
				"rio.cattle.io/stack":   stack.Stack.Name,
				"rio.cattle.io/project": stack.Stack.Namespace,
			},
		},
	}
}

func destWeightForExternalService(d v1.WeightedDestination, svc *v1.ExternalService) *v1alpha3.DestinationWeight {
	if d.Port == 0 {
		d.Port = 80
	}
	if d.Weight == 0 {
		d.Weight = 100
	}
	return &v1alpha3.DestinationWeight{
		Destination: &v1alpha3.Destination{
			Host: svc.Spec.Target,
			Port: &v1alpha3.PortSelector{
				Port: &v1alpha3.PortSelector_Number{
					Number: uint32(d.Port),
				},
			},
		},
		Weight: int32(d.Weight),
	}
}

func destWeightForService(d v1.WeightedDestination) *v1alpha3.DestinationWeight {
	if d.Revision == "" {
		d.Revision = "v0"
	}
	if d.Port == 0 {
		d.Port = 80
	}
	if d.Weight == 0 {
		d.Weight = 100
	}
	return &v1alpha3.DestinationWeight{
		Destination: &v1alpha3.Destination{
			Host:   d.Service,
			Subset: d.Revision,
			Port: &v1alpha3.PortSelector{
				Port: &v1alpha3.PortSelector_Number{
					Number: uint32(d.Port),
				},
			},
		},
		Weight: int32(d.Weight),
	}
}
