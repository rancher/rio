package populate

import (
	"fmt"
	"hash/adler32"
	"strconv"
	"strings"
	"time"

	"github.com/rancher/rio/modules/gateway/controllers/service/populate"

	"github.com/knative/pkg/apis/istio/common/v1alpha1"
	"github.com/knative/pkg/apis/istio/v1alpha3"
	"github.com/rancher/mapper/slice"
	"github.com/rancher/rio/modules/istio/pkg/domains"
	"github.com/rancher/rio/modules/istio/pkg/parse"
	projectv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/kv"
	"github.com/rancher/wrangler/pkg/objectset"
	v1alpha3type "istio.io/api/networking/v1alpha3"
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
		Hosts: []string{routeSet.Name, domains.GetExternalDomain(routeSet.Name, routeSet.Namespace, clusterDomain.Status.ClusterDomain)},
	}

	pb := v1.ContainerPort{
		Port:       8089,
		TargetPort: 8089,
		Protocol:   v1.ProtocolHTTP,
	}
	for _, publicDomain := range routeSet.Status.PublicDomains {
		spec.Hosts = append(spec.Hosts, publicDomain)
		ds := []populate.Dest{
			{
				Host:   fmt.Sprintf("cm-acme-http-solver-%d", adler32.Checksum([]byte(publicDomain))),
				Subset: constants.AcmeVersion,
				Weight: 100,
			},
		}
		_, route := populate.NewRoute(true, systemNamespace, domains.GetPublicGateway(systemNamespace), true, pb, ds, false, false, nil)
		route.Match[0].URI = &v1alpha1.StringMatch{
			Prefix: "/.well-known/acme-challenge/",
		}
		route.Match[0].Authority = &v1alpha1.StringMatch{
			Prefix: publicDomain,
		}
		spec.HTTP = append(spec.HTTP, route)
	}

	// populate http routing
	for index, routeSpec := range routeSet.Spec.Routes {
		httpRoute := v1alpha3.HTTPRoute{
			Headers: &v1alpha3.Headers{
				Request: routeSpec.Headers,
			},
		}

		// populate destinations
		for _, dest := range routeSpec.To {
			if httpRoute.Headers.Request == nil {
				httpRoute.Headers.Request = &v1alpha3.HeaderOperations{}
			}
			if httpRoute.Headers.Request.Set == nil {
				httpRoute.Headers.Request.Set = map[string]string{}
			}
			port := 80
			if dest.Port != nil {
				port = int(*dest.Port)
			}

			if constants.ServiceMeshMode == constants.ServiceMeshModeLinkerd {
				// https://linkerd.io/2/tasks/using-ingress/ In linkerd we override header to make sure traffic is sent to the desired target. For router it is going to be ${name}-${route-index}.${namespace}.svc.cluster.local
				httpRoute.Headers.Request.Set["l5d-dst-override"] = fmt.Sprintf("%s-%v.%s.svc.cluster.local:%v", routeSet.Name, index, routeSet.Namespace, port)
				if !slice.ContainsString(httpRoute.Headers.Request.Remove, "l5d-remote-ip") || !slice.ContainsString(httpRoute.Headers.Request.Remove, "l5d-remote-ip") {
					httpRoute.Headers.Request.Remove = append(httpRoute.Headers.Request.Remove, []string{
						"l5d-remote-ip",
						"l5d-server-id",
					}...)
				}
			}

			if dest.Destination.Namespace == "" {
				dest.Destination.Namespace = routeSet.Namespace
			}
			if esvc, ok := externalServiceMap[dest.Service]; ok {
				httpRoute.Route = append(httpRoute.Route, destWeightForExternalService(dest, esvc))
			} else if _, ok := routesetMap[dest.Service]; ok {
				httpRoute.Route = append(httpRoute.Route, destWeightForRouteset(dest))
				routeSpec.Rewrite = &v1.Rewrite{
					Host: fmt.Sprintf("%s-%s.%s", dest.Destination.Service, dest.Destination.Namespace, clusterDomain.Status.ClusterDomain),
				}
				localhostServiceEntry(os, routeSet.Namespace)
			} else {
				httpRoute.Route = append(httpRoute.Route, destWeightForService(dest, routeSet.Namespace))
			}
		}

		// populate matches
		for _, match := range routeSpec.Matches {
			httpMatch := v1alpha3.HTTPMatchRequest{
				URI:    populateStringMatch(match.Path),
				Scheme: populateStringMatch(match.Scheme),
				Method: populateStringMatch(match.Method),
			}
			if match.Port != nil {
				httpMatch.Port = uint32(*match.Port)
			}
			if len(match.Cookies) != 0 || len(match.Headers) != 0 {
				httpMatch.Headers = map[string]v1alpha1.StringMatch{}
			}
			// todo: looks like istio doesn't support multiple headers, take the first one
			for name, cookie := range match.Cookies {
				match := populateStringMatch(&cookie)
				if match != nil {
					r := v1alpha1.StringMatch{}
					if match.Exact != "" {
						r.Exact = fmt.Sprintf("%s=%s", name, match.Exact)
					} else if match.Prefix != "" {
						r.Prefix = fmt.Sprintf("%s=%s", name, match.Prefix)
					} else if match.Regex != "" {
						r.Regex = fmt.Sprintf("%s=%s", name, match.Regex)
					} else if match.Suffix != "" {
						r.Suffix = fmt.Sprintf("%s=%s", name, match.Suffix)
					}
					httpMatch.Headers["Cookie"] = r
				}
				break
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
			httpPort, _ := strconv.Atoi(constants.DefaultHTTPOpenPort)
			httpsPort, _ := strconv.Atoi(constants.DefaultHTTPSOpenPort)
			httpRoute.Match = []v1alpha3.HTTPMatchRequest{
				{
					Gateways: []string{privateGw, domains.GetPublicGateway(systemNamespace)},
					Port:     uint32(httpPort),
				},
				{
					Gateways: []string{privateGw, domains.GetPublicGateway(systemNamespace)},
					Port:     uint32(httpsPort),
				},
			}
		}

		if routeSpec.Redirect != nil {
			httpRoute.Redirect = &v1alpha3.HTTPRedirect{
				URI:       routeSpec.Redirect.Path,
				Authority: routeSpec.Redirect.Host,
			}
		}

		if routeSpec.Rewrite != nil {
			httpRoute.Rewrite = &v1alpha3.HTTPRewrite{
				URI:       routeSpec.Rewrite.Path,
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
				Abort: populateHTTPAbort(routeSpec.Fault),
			}
		}

		if routeSpec.TimeoutMillis != nil && *routeSpec.TimeoutMillis > 0 {
			httpRoute.Timeout = (time.Millisecond * time.Duration(*routeSpec.TimeoutMillis)).String()
		}

		if routeSpec.Mirror != nil {
			httpRoute.Mirror = &v1alpha3.Destination{
				Host:   fmt.Sprintf("%s.%s.svc.cluster.local", routeSpec.Mirror.Service, routeSpec.Mirror.Namespace),
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

		spec.HTTP = append(spec.HTTP, httpRoute)
	}

	// set port to 80 for virtual services that are created from gateway
	vs := constructors.NewVirtualService(routeSet.Namespace, routeSet.Name, v1alpha3.VirtualService{})
	vs.Spec = spec

	return vs
}

func populateHTTPAbort(fault *v1.Fault) *v1alpha3.InjectAbort {
	abort := &v1alpha3.InjectAbort{
		Percent:    fault.Percentage,
		HTTPStatus: fault.Abort.HTTPStatus,
	}
	return abort
}

func populateStringMatch(match *v1.StringMatch) *v1alpha1.StringMatch {
	if match == nil {
		return nil
	}
	m := &v1alpha1.StringMatch{}
	switch {
	case match.Exact != "":
		m.Exact = match.Exact
	case match.Prefix != "":
		m.Prefix = match.Prefix
	case match.Regexp != "":
		m.Regex = match.Regexp
	default:
		return nil
	}
	return m
}

func destWeightForService(d v1.WeightedDestination, defaultNamespace string) v1alpha3.HTTPRouteDestination {
	if d.Port == nil {
		d.Port = &[]uint32{80}[0]
	}
	if d.Weight == 0 {
		d.Weight = 100
	}
	if d.Namespace == "" {
		d.Namespace = defaultNamespace
	}
	host := fmt.Sprintf("%s.%s.svc.cluster.local", d.Service, d.Namespace)
	return v1alpha3.HTTPRouteDestination{
		Destination: v1alpha3.Destination{
			Host:   host,
			Subset: d.Revision,
			Port: v1alpha3.PortSelector{
				Number: *d.Port,
			},
		},
		Weight: d.Weight,
	}
}

func destWeightForExternalService(d v1.WeightedDestination, esvc *v1.ExternalService) v1alpha3.HTTPRouteDestination {
	if d.Port == nil {
		d.Port = &[]uint32{80}[0]
	}
	switch {
	case esvc.Spec.FQDN != "":
		// ignore error as it should be validated somewhere else
		u, _ := parse.TargetURL(esvc.Spec.FQDN)
		d.Service = u.Host
	case esvc.Spec.Service != "":
		stackName, serviceName := kv.Split(esvc.Spec.Service, "/")
		d.Service = fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, stackName)
	case len(esvc.Spec.IPAddresses) > 0:
		d.Service = fmt.Sprintf("%s.%s.svc.cluster.local", esvc.Name, esvc.Namespace)
	}
	return v1alpha3.HTTPRouteDestination{
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

func destWeightForRouteset(d v1.WeightedDestination) v1alpha3.HTTPRouteDestination {
	if d.Port == nil {
		d.Port = &[]uint32{80}[0]
	}
	if d.Weight == 0 {
		d.Weight = 100
	}
	return v1alpha3.HTTPRouteDestination{
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
