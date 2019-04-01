package populate

import (
	"fmt"
	"strings"
	"time"

	"github.com/knative/pkg/apis/istio/common/v1alpha1"
	"github.com/knative/pkg/apis/istio/v1alpha3"
	"github.com/rancher/rio/modules/istio/pkg/domains"
	projectv1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/upstream2/exclude/features/routing/controllers/externalservice/populate"
	"github.com/rancher/wrangler/pkg/kv"
	"github.com/rancher/wrangler/pkg/objectset"
	v1alpha3type "istio.io/api/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	privateGw = "mesh"
)

func VirtualServices(systemNamespace string, clusterDomain *projectv1.ClusterDomain, routeSet *v1.Router, externalServiceMap map[string]*v1.ExternalService, routesetMap map[string]*v1.Router, os *objectset.ObjectSet) error {
	vs := virtualServiceFromRoutesets(systemNamespace, clusterDomain, routeSet, externalServiceMap, routesetMap, os)
	if vs != nil {
		os.Add(vs)
	}

	return nil
}

func virtualServiceFromRoutesets(systemNamespace string, clusterDomain *projectv1.ClusterDomain, routeSet *v1.Router, externalServiceMap map[string]*v1.ExternalService, routesetMap map[string]*v1.Router, os *objectset.ObjectSet) *v1alpha3.VirtualService {
	spec := v1alpha3.VirtualServiceSpec{
		Gateways: []string{
			privateGw,
			domains.GetPublicGateway(systemNamespace),
		},
		Hosts: []string{domains.GetExternalDomain(routeSet.Name, routeSet.Namespace, clusterDomain.Status.ClusterDomain)},
	}

	// populate http routing
	for _, routeSpec := range routeSet.Spec.Routes {
		httpRoute := v1alpha3.HTTPRoute{}
		// populate destinations
		for _, dest := range routeSpec.To {
			if dest.Destination.Stack == "" {
				dest.Destination.Stack = routeSet.Namespace
			}
			if esvc, ok := externalServiceMap[dest.Service]; ok {
				httpRoute.Route = append(httpRoute.Route, destWeightForExternalService(dest, esvc))
			} else if _, ok := routesetMap[dest.Service]; ok {
				httpRoute.Route = append(httpRoute.Route, destWeightForRouteset(dest))
				routeSpec.Rewrite = &v1.Rewrite{
					Host: fmt.Sprintf("%s-%s.%s", dest.Destination.Service, dest.Destination.Stack, clusterDomain.Status.ClusterDomain),
				}
				localhostServiceEntry(os, routeSet.Namespace)
			} else {
				httpRoute.Route = append(httpRoute.Route, destWeightForService(dest, routeSet.Namespace))
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
					Gateways: []string{privateGw, domains.GetPublicGateway(systemNamespace)},
					Port:     80,
				},
				{
					Gateways: []string{privateGw, domains.GetPublicGateway(systemNamespace)},
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
				Host:   domains.GetExternalDomain(routeSpec.Mirror.Service, routeSpec.Mirror.Stack, clusterDomain.Status.ClusterDomain),
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

		spec.Http = append(spec.Http, httpRoute)
	}

	// set port to 80 for virtual services that are created from gateway
	vs := newVirtualServiceGeneric(routeSet.Name, routeSet.Namespace)
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

func newVirtualServiceGeneric(name, namespace string) *v1alpha3.VirtualService {
	return constructors.NewVirtualService(namespace, name, v1alpha3.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
			Labels:      map[string]string{},
		},
	})
}

func destWeightForService(d v1.WeightedDestination, defaultNamespace string) v1alpha3.DestinationWeight {
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
		d.Stack = defaultNamespace
	}
	return v1alpha3.DestinationWeight{
		Destination: v1alpha3.Destination{
			Host:   fmt.Sprintf("%s.%s.svc.cluster.local", d.Service, d.Stack),
			Subset: d.Revision,
			Port: v1alpha3.PortSelector{
				Number: *d.Port,
			},
		},
		Weight: d.Weight,
	}
}

func destWeightForExternalService(d v1.WeightedDestination, esvc *v1.ExternalService) v1alpha3.DestinationWeight {
	if d.Port == nil {
		d.Port = &[]uint32{80}[0]
	}
	if esvc.Spec.FQDN != "" {
		// ignore error as it should be validated somewhere else
		u, _ := populate.ParseTargetUrl(esvc.Spec.FQDN)
		d.Service = u.Host
	} else if esvc.Spec.Service != "" {
		stackName, serviceName := kv.Split(esvc.Spec.Service, "/")
		d.Service = fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, stackName)
	} else if len(esvc.Spec.IPAddresses) > 0 {
		d.Service = fmt.Sprintf("%s.%s.svc.cluster.local", esvc.Name, esvc.Namespace)
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
	se := constructors.NewServiceEntry(namespace, "localhost", v1alpha3.ServiceEntry{
		Spec: v1alpha3.ServiceEntrySpec{
			Hosts:      []string{"localhost.localhost"},
			Location:   int32(v1alpha3type.ServiceEntry_MESH_EXTERNAL),
			Resolution: int32(v1alpha3type.ServiceEntry_DNS),
			Ports: []v1alpha3.Port{
				{
					Protocol: v1alpha3.PortProtocol(strings.ToUpper("http")),
					Number:   80,
					Name:     "http-80",
				},
			},
		},
	})
	os.Add(se)
}
