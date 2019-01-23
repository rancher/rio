package populate

import (
	"fmt"
	"strings"
	"time"

	"github.com/knative/pkg/apis/istio/common/v1alpha1"
	"github.com/knative/pkg/apis/istio/v1alpha3"
	"github.com/rancher/norman/pkg/kv"
	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/features/routing/controllers/externalservice/populate"
	"github.com/rancher/rio/features/routing/pkg/domains"
	"github.com/rancher/rio/pkg/namespace"
	v1alpha3client "github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3"
	v1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	v1alpha3type "istio.io/api/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	privateGw              = "mesh"
	PublicDomainAnnotation = "rio.cattle.io/publicDomain"
)

func VirtualServices(stack *v1.Stack, routeSet *v1.RouteSet, externalServiceMap map[string]*v1.ExternalService, routesetMap map[string]*v1.RouteSet, os *objectset.ObjectSet) error {
	vs := vsFromRoutesets(stack, routeSet, externalServiceMap, routesetMap, os)
	if vs != nil {
		os.Add(vs)
	}

	return nil
}

func vsFromRoutesets(stack *v1.Stack, routeSet *v1.RouteSet, externalServiceMap map[string]*v1.ExternalService, routesetMap map[string]*v1.RouteSet, os *objectset.ObjectSet) *v1alpha3client.VirtualService {
	spec := v1alpha3.VirtualServiceSpec{
		Gateways: []string{
			privateGw,
			domains.GetPublicGateway(),
		},
		Hosts: []string{
			routeSet.Name,
			domains.GetExternalDomain(routeSet.Name, stack.Name, stack.Namespace)},
	}

	// populate http routing
	for _, routeSpec := range routeSet.Spec.Routes {
		httpRoute := v1alpha3.HTTPRoute{}
		// populate destinations
		for _, dest := range routeSpec.To {
			if esvc, ok := externalServiceMap[dest.Service]; ok {
				httpRoute.Route = append(httpRoute.Route, destWeightForExternalService(dest, esvc, stack))
			} else if _, ok := routesetMap[dest.Service]; ok {
				httpRoute.Route = append(httpRoute.Route, destWeightForRouteset(dest))
				routeSpec.Rewrite = &v1.Rewrite{
					Host: fmt.Sprintf("%s.%s.svc.cluster.local", dest.Destination.Service, namespace.StackNamespace(stack.Namespace, stack.Name)),
				}
				localhostServiceEntry(os, routeSet.Namespace)
			} else {
				httpRoute.Route = append(httpRoute.Route, destWeightForService(dest, stack))
			}
		}

		// populate matches
		for _, match := range routeSpec.Matches {
			httpMatch := v1alpha3.HTTPMatchRequest{
				Uri:    populateStringMatch(match.Path),
				Scheme: populateStringMatch(match.Scheme),
				Method: populateStringMatch(match.Method),
			}
			if match.Port != nil {
				httpMatch.Port = uint32(*match.Port)
			}
			httpMatch.Headers = map[string]v1alpha1.StringMatch{}
			for name, cookie := range match.Cookies {
				match := populateStringMatch(&cookie)
				if match != nil {
					httpMatch.Headers[name] = *match
				}
			}
			for name, value := range match.Headers {
				match := populateStringMatch(&value)
				if match != nil {
					httpMatch.Headers[name] = *match
				}
			}
			httpRoute.Match = append(httpRoute.Match, httpMatch)
		}
		if len(httpRoute.Match) == 0 {
			httpRoute.Match = []v1alpha3.HTTPMatchRequest{
				{
					Gateways: []string{privateGw, domains.GetPublicGateway()},
					Port:     80,
				},
				{
					Gateways: []string{privateGw, domains.GetPublicGateway()},
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
		if routeSpec.Fault != nil && routeSpec.Fault.DelayMillis > 0 {
			httpRoute.Fault = &v1alpha3.HTTPFaultInjection{
				Delay: &v1alpha3.InjectDelay{
					Percent:    routeSpec.Fault.Percentage,
					FixedDelay: (time.Millisecond * time.Duration(routeSpec.Fault.DelayMillis)).String(),
				},
				Abort: populateHttpAbort(routeSpec.Fault),
			}
		}

		if routeSpec.TimeoutMillis != nil && *routeSpec.TimeoutMillis > 0 {
			httpRoute.Timeout = (time.Millisecond * time.Duration(*routeSpec.TimeoutMillis)).String()
		}

		if routeSpec.Mirror != nil {
			httpRoute.Mirror = &v1alpha3.Destination{
				Host:   domains.GetExternalDomain(routeSpec.Mirror.Service, routeSpec.Mirror.Stack, stack.Namespace),
				Subset: routeSpec.Mirror.Revision,
			}
			if routeSpec.Mirror.Port != nil {
				httpRoute.Mirror.Port = v1alpha3.PortSelector{
					Number: *routeSpec.Mirror.Port,
				}
			}
		}

		if routeSpec.Retry != nil {
			httpRoute.Retries = &v1alpha3.HTTPRetry{
				Attempts:      routeSpec.Retry.Attempts,
				PerTryTimeout: (time.Millisecond * time.Duration(routeSpec.Retry.TimeoutMillis)).String(),
			}
		}

		if routeSet.Annotations[PublicDomainAnnotation] != "" {
			spec.Hosts = append(spec.Hosts, strings.Split(routeSet.Annotations[PublicDomainAnnotation], ",")...)
		}

		spec.Http = append(spec.Http, httpRoute)
	}

	// set port to 80 for virtual services that are created from gateway
	vs := newVirtualServiceGeneric(stack, routeSet.Name, routeSet.Namespace)
	vs.Annotations["rio.cattle.io/ports"] = "80"
	vs.Spec = spec

	return vs
}

func populateHttpAbort(fault *v1.Fault) *v1alpha3.InjectAbort {
	abort := &v1alpha3.InjectAbort{
		Percent: fault.Percentage,
	}
	if fault.Abort.GRPCStatus != "" {
		abort.GrpcStatus = fault.Abort.GRPCStatus
	} else if fault.Abort.HTTP2Status != "" {
		abort.Http2Error = fault.Abort.HTTP2Status
	} else if fault.Abort.HTTPStatus != 0 {
		abort.HttpStatus = fault.Abort.HTTPStatus
	}
	return abort
}

func populateStringMatch(match *v1.StringMatch) *v1alpha1.StringMatch {
	if match == nil {
		return nil
	}
	m := &v1alpha1.StringMatch{}
	if match.Exact != "" {
		m.Exact = match.Exact
	} else if match.Prefix != "" {
		m.Prefix = match.Prefix
	} else if match.Regexp != "" {
		m.Regex = match.Regexp
	} else {
		return nil
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

func newVirtualServiceGeneric(stack *v1.Stack, name, namespace string) *v1alpha3client.VirtualService {
	return v1alpha3client.NewVirtualService(namespace, name, v1alpha3client.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
			Labels: map[string]string{
				"rio.cattle.io/stack":   stack.Name,
				"rio.cattle.io/project": stack.Namespace,
			},
		},
	})
}

func destWeightForService(d v1.WeightedDestination, stack *v1.Stack) v1alpha3.DestinationWeight {
	if d.Revision == "" {
		d.Revision = "v0"
	}
	if d.Port == nil {
		d.Port = &[]uint32{80}[0]
	}
	if d.Weight == 0 {
		d.Weight = 100
	}
	if d.Stack == "" {
		d.Stack = stack.Name
	}
	return v1alpha3.DestinationWeight{
		Destination: v1alpha3.Destination{
			Host:   fmt.Sprintf("%s.%s.svc.cluster.local", d.Service, namespace.StackNamespace(stack.Namespace, d.Stack)),
			Subset: d.Revision,
			Port: v1alpha3.PortSelector{
				Number: *d.Port,
			},
		},
		Weight: d.Weight,
	}
}

func destWeightForExternalService(d v1.WeightedDestination, esvc *v1.ExternalService, stack *v1.Stack) v1alpha3.DestinationWeight {
	if d.Port == nil {
		d.Port = &[]uint32{80}[0]
	}
	if esvc.Spec.FQDN != "" {
		// ignore error as it should be validated somewhere else
		u, _ := populate.ParseTargetUrl(esvc.Spec.FQDN)
		d.Service = u.Host
	} else if esvc.Spec.Service != "" {
		stackName, serviceName := kv.Split(esvc.Spec.Service, "/")
		d.Service = fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, namespace.StackNamespace(stack.Namespace, stackName))
	} else if len(esvc.Spec.IPAddresses) > 0 {
		d.Service = fmt.Sprintf("%s.%s.svc.cluster.local", esvc.Name, namespace.StackNamespace(stack.Namespace, stack.Name))
	}
	return v1alpha3.DestinationWeight{
		Destination: v1alpha3.Destination{
			Host:   d.Service,
			Subset: d.Revision,
			Port: v1alpha3.PortSelector{
				Number: *d.Port,
			},
		},
		Weight: d.Weight,
	}
}

func destWeightForRouteset(d v1.WeightedDestination) v1alpha3.DestinationWeight {
	if d.Port == nil {
		d.Port = &[]uint32{80}[0]
	}
	if d.Weight == 0 {
		d.Weight = 100
	}
	return v1alpha3.DestinationWeight{
		Destination: v1alpha3.Destination{
			Host: "localhost.localhost",
			Port: v1alpha3.PortSelector{
				Number: *d.Port,
			},
		},
		Weight: d.Weight,
	}
}

func localhostServiceEntry(os *objectset.ObjectSet, namespace string) {
	se := v1alpha3client.NewServiceEntry(namespace, "localhost", v1alpha3client.ServiceEntry{
		Spec: v1alpha3client.ServiceEntrySpec{
			Hosts:      []string{"localhost.localhost"},
			Location:   v1alpha3type.ServiceEntry_MESH_EXTERNAL,
			Resolution: v1alpha3type.ServiceEntry_DNS,
			Ports: []v1alpha3client.Port{
				{
					Protocol: strings.ToUpper("http"),
					Number:   80,
					Name:     "http-80",
				},
			},
		},
	})
	os.Add(se)
}
