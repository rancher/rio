package populate

import (
	"fmt"
	"hash/adler32"
	"strconv"

	"github.com/knative/pkg/apis/istio/common/v1alpha1"
	"github.com/knative/pkg/apis/istio/v1alpha3"
	"github.com/rancher/rio/modules/istio/pkg/domains"
	projectv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/rio/pkg/serviceset"
	"github.com/rancher/wrangler/pkg/objectset"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	privateGw = "mesh"

	ProbeHeaderName = "K-Network-Probe"
	// RevisionHeaderName is the header key for revision name.
	RevisionHeaderName string = "knative-serving-revision"
	// RevisionHeaderNamespace is the header key for revision's namespace.
	RevisionHeaderNamespace string = "knative-serving-namespace"
)

func VirtualServices(namespace string, clusterDomain *projectv1.ClusterDomain, service *v1.Service, os *objectset.ObjectSet) error {
	os.Add(virtualServiceFromService(namespace, clusterDomain, service)...)
	return nil
}

func httpRoutes(systemNamespace string, service *v1.Service, dests []Dest) ([]v1alpha3.HTTPRoute, bool) {
	external := false
	var result []v1alpha3.HTTPRoute

	// add https challenge match
	pb := v1.ContainerPort{
		Port:       8089,
		TargetPort: 8089,
		Protocol:   v1.ProtocolHTTP,
	}
	for _, publicDomain := range service.Status.PublicDomains {
		ds := []Dest{
			{
				Host:   fmt.Sprintf("cm-acme-http-solver-%d", adler32.Checksum([]byte(publicDomain))),
				Subset: "latest",
				Weight: 100,
			},
		}
		_, route := newRoute(systemNamespace, domains.GetPublicGateway(systemNamespace), true, pb, ds, false, false, nil)
		route.Match[0].URI = &v1alpha1.StringMatch{
			Prefix: "/.well-known/acme-challenge/",
		}
		route.Match[0].Authority = &v1alpha1.StringMatch{
			Prefix: publicDomain,
		}
		result = append(result, route)
	}

	autoscale := false
	if service.Spec.MaxScale != nil && service.Spec.Concurrency != nil && service.Spec.MinScale != nil && *service.Spec.MaxScale != *service.Spec.MinScale {
		autoscale = true
	}
	if autoscale && service.Status.ObservedScale != nil && *service.Status.ObservedScale == 0 {
		for _, port := range service.Spec.Ports {
			if port.InternalOnly {
				continue
			}
			publicPort, route := newRoute(systemNamespace, domains.GetPublicGateway(systemNamespace), !port.InternalOnly, port, dests, true, false, service)
			if publicPort != "" {
				route.Match = []v1alpha3.HTTPMatchRequest{
					{
						Headers: map[string]v1alpha1.StringMatch{
							ProbeHeaderName: {
								Exact: "queue",
							},
						},
					},
				}
				result = append(result, route)
			}
		}
	}

	for _, port := range service.Spec.Ports {
		if port.InternalOnly {
			continue
		}
		publicPort, route := newRoute(systemNamespace, domains.GetPublicGateway(systemNamespace), !port.InternalOnly, port, dests, true, autoscale, service)
		if publicPort != "" {
			external = true
			result = append(result, route)
		}
	}

	return result, external
}
func newRoute(systemNamespace, externalGW string, published bool, portBinding v1.ContainerPort, dests []Dest, appendHTTPS bool, autoscale bool, svc *v1.Service) (string, v1alpha3.HTTPRoute) {
	route := v1alpha3.HTTPRoute{}

	if portBinding.Protocol == "" {
		portBinding.Protocol = v1.ProtocolHTTP
	}

	if !isProtocolSupported(portBinding.Protocol) {
		return "", route
	}

	gw := []string{privateGw}
	if published {
		gw = append(gw, externalGW)
	}

	httpPort, _ := strconv.ParseUint(constants.DefaultHTTPOpenPort, 10, 64)
	httpsPort, _ := strconv.ParseUint(constants.DefaultHTTPSOpenPort, 10, 64)
	matches := []v1alpha3.HTTPMatchRequest{
		{
			Port:     uint32(httpPort),
			Gateways: gw,
		},
	}
	if appendHTTPS {
		matches = append(matches,
			v1alpha3.HTTPMatchRequest{
				Port:     uint32(httpsPort),
				Gateways: gw,
			})
	}
	matches = append(matches, v1alpha3.HTTPMatchRequest{
		Port:     uint32(portBinding.Port),
		Gateways: []string{privateGw},
	})
	route.Match = matches

	if autoscale && deployIsZero(svc) {
		if route.Headers == nil {
			route.Headers = &v1alpha3.Headers{
				Request: &v1alpha3.HeaderOperations{
					Add: map[string]string{
						RevisionHeaderName:      svc.Name,
						RevisionHeaderNamespace: svc.Namespace,
					},
				},
			}
		}
	}

	for _, dest := range dests {
		if autoscale && deployIsZero(svc) {
			route.Route = append(route.Route, v1alpha3.HTTPRouteDestination{
				Destination: v1alpha3.Destination{
					Host: fmt.Sprintf("%s.%s.svc.cluster.local", "activator", systemNamespace),
					Port: v1alpha3.PortSelector{
						Number: 8012,
					},
				},
				Weight: 100,
			})
		} else {
			ns := systemNamespace
			if svc != nil {
				ns = svc.Namespace
			}
			route.Route = append(route.Route, v1alpha3.HTTPRouteDestination{
				Destination: v1alpha3.Destination{
					Host:   fmt.Sprintf("%s.%s.svc.cluster.local", dest.Host, ns),
					Subset: dest.Subset,
					Port: v1alpha3.PortSelector{
						Number: uint32(portBinding.Port),
					},
				},
				Weight: dest.Weight,
			})
		}
	}

	sourcePort := httpPort
	if portBinding.Protocol == "https" {
		sourcePort = httpsPort
	}
	return fmt.Sprintf("%v/%s", sourcePort, portBinding.Protocol), route
}

func deployIsZero(service *riov1.Service) bool {
	if service.Status.DeploymentStatus != nil && service.Status.DeploymentStatus.Replicas == 0 {
		return true
	}
	return false
}

type Dest struct {
	Host, Subset string
	Weight       int
}

func DestsForService(service *serviceset.ServiceSet) []Dest {
	var result []Dest
	for _, rev := range service.Revisions {
		app, ver := services.AppAndVersion(rev)
		weight := rev.Spec.ServiceRevision.Weight
		if rev.Status.WeightOverride != nil {
			weight = *rev.Status.WeightOverride
		}
		result = append(result, Dest{
			Host:   app,
			Weight: weight,
			Subset: ver,
		})
	}

	return result
}

func virtualServiceFromService(namespace string, clusterDomain *projectv1.ClusterDomain, service *riov1.Service) []runtime.Object {
	var result []runtime.Object

	// virtual service for each revision
	app, version := services.AppAndVersion(service)
	revVs := VirtualServiceFromSpec(false, namespace, app+"-"+version, service.Namespace, clusterDomain, service, Dest{
		Host:   app,
		Subset: version,
		Weight: 100,
	})
	if revVs != nil {
		result = append(result, revVs)
	}

	return result
}

func isProtocolSupported(protocol v1.Protocol) bool {
	if protocol == v1.ProtocolHTTP || protocol == v1.ProtocolHTTP2 || protocol == v1.ProtocolGRPC || protocol == v1.ProtocolTCP {
		return true
	}
	return false
}

func VirtualServiceFromSpec(aggregated bool, systemNamespace string, name, namespace string, clusterDomain *projectv1.ClusterDomain, service *v1.Service, dests ...Dest) *v1alpha3.VirtualService {
	routes, external := httpRoutes(systemNamespace, service, dests)
	if len(routes) == 0 {
		return nil
	}

	if clusterDomain.Status.ClusterDomain == "" {
		external = false
	}

	vs := constructors.NewVirtualService(namespace, name, v1alpha3.VirtualService{})
	spec := v1alpha3.VirtualServiceSpec{
		Gateways: []string{privateGw},
		HTTP:     routes,
	}
	if aggregated {
		spec.Hosts = []string{name}
	}

	for _, publicDomain := range service.Status.PublicDomains {
		if publicDomain == "" {
			continue
		}
		spec.Hosts = append(spec.Hosts, publicDomain)
	}

	if external {
		externalGW := domains.GetPublicGateway(systemNamespace)
		externalHost := domains.GetExternalDomain(name, namespace, clusterDomain.Status.ClusterDomain)
		spec.Gateways = append(spec.Gateways, externalGW)
		spec.Hosts = append(spec.Hosts, externalHost)
	}

	vs.Spec = spec
	return vs
}
