package populate

import (
	"fmt"
	"strconv"

	"github.com/knative/pkg/apis/istio/v1alpha3"
	"github.com/rancher/rio/modules/istio/pkg/domains"
	projectv1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/pkg/serviceset"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/wrangler/pkg/objectset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	privateGw          = "mesh"
	RioNameHeader      = "X-Rio-ServiceName"
	RioNamespaceHeader = "X-Rio-Namespace"
	RioPortHeader      = "X-Rio-ServicePort"
)

func virtualServices(namespace string, clusterDomain *projectv1.ClusterDomain, services []*v1.Service, service *v1.Service, os *objectset.ObjectSet) error {
	serviceSets, err := serviceset.CollectionServices(services)
	if err != nil {
		return err
	}

	serviceSet, ok := serviceSets[service.Name]
	if !ok {
		return nil
	}

	os.Add(virtualServiceFromService(namespace, service.Name, clusterDomain, serviceSet)...)
	return nil
}

func httpRoutes(systemNamespace string, service *v1.Service, dests []Dest) ([]v1alpha3.HTTPRoute, bool) {
	external := false
	var result []v1alpha3.HTTPRoute

	enableAutoScale := service.Spec.AutoScale != nil
	for _, port := range service.Spec.Ports {
		publicPort, route := newRoute(domains.GetPublicGateway(systemNamespace), port.Publish, port, dests, true, enableAutoScale, service)
		if publicPort != "" {
			external = true
			result = append(result, route)
		}
	}

	return result, external
}

func newRoute(externalGW string, published bool, portBinding v1.ServicePort, dests []Dest, appendHttps bool, autoscale bool, svc *v1.Service) (string, v1alpha3.HTTPRoute) {
	route := v1alpha3.HTTPRoute{}

	if !isProtocolSupported(portBinding.Protocol) {
		return "", route
	}

	gw := []string{privateGw}
	if published {
		gw = append(gw, externalGW)
	}

	httpPort, _ := strconv.ParseUint(settings.DefaultHTTPOpenPort, 10, 64)
	httpsPort, _ := strconv.ParseUint(settings.DefaultHTTPSOpenPort, 10, 64)
	matches := []v1alpha3.HTTPMatchRequest{
		{
			Port:     uint32(httpPort),
			Gateways: gw,
		},
	}
	if appendHttps {
		matches = append(matches,
			v1alpha3.HTTPMatchRequest{
				Port:     uint32(httpsPort),
				Gateways: gw,
			})
	}
	route.Match = matches

	if autoscale {
		if route.AppendHeaders == nil {
			route.AppendHeaders = map[string]string{}
		}
		route.AppendHeaders[RioNameHeader] = svc.Name
		route.AppendHeaders[RioNamespaceHeader] = svc.Namespace
		route.AppendHeaders[RioPortHeader] = portBinding.TargetPort.String()
		route.Retries = &v1alpha3.HTTPRetry{
			PerTryTimeout: "1m",
			Attempts:      3,
		}
	}

	for _, dest := range dests {
		if autoscale && svc.Spec.Scale == 0 {
			route.Route = append(route.Route, v1alpha3.DestinationWeight{
				Destination: v1alpha3.Destination{
					Host: fmt.Sprintf("%s.%s.svc.cluster.local", "gateway", settings.AutoScaleStack),
					Port: v1alpha3.PortSelector{
						Number: 80,
					},
				},
			})
		} else {
			route.Route = append(route.Route, v1alpha3.DestinationWeight{
				Destination: v1alpha3.Destination{
					Host:   dest.Host,
					Subset: dest.Subset,
					Port: v1alpha3.PortSelector{
						Number: uint32(portBinding.TargetPort.IntValue()),
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

type Dest struct {
	Host, Subset string
	Weight       int
}

func DestsForService(namespace, name string, service *serviceset.ServiceSet) []Dest {
	latestWeight := 100
	result := []Dest{
		{
			Host:   fmt.Sprintf("%s.%s.svc.cluster.local", name, namespace),
			Subset: service.Service.Spec.Revision.Version,
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

		result = append(result, Dest{
			Host:   fmt.Sprintf("%s.%s.svc.cluster.local", rev.Name, service.Service.Namespace),
			Weight: weight,
			Subset: rev.Spec.Revision.Version,
		})
	}

	result[0].Weight = latestWeight
	if result[0].Weight == 0 && len(result) > 1 {
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

func virtualServiceFromService(namespace string, name string, clusterDomain *projectv1.ClusterDomain, service *serviceset.ServiceSet) []runtime.Object {
	var result []runtime.Object

	serviceVS := VirtualServiceFromSpec(namespace, name, service.Service.Namespace, clusterDomain, service.Service, DestsForService(namespace, name, service)...)
	if serviceVS != nil {
		result = append(result, serviceVS)
	}

	for _, rev := range service.Revisions {
		revVs := VirtualServiceFromSpec(namespace, rev.Name, service.Service.Namespace, clusterDomain, rev, Dest{
			Host:   rev.Name,
			Subset: rev.Spec.Revision.Version,
			Weight: 100,
		})
		if revVs != nil {
			result = append(result, revVs)
		}
	}

	return result
}

func VirtualServiceFromSpec(systemNamespace string, name, namespace string, clusterDomain *projectv1.ClusterDomain, service *v1.Service, dests ...Dest) *v1alpha3.VirtualService {
	routes, external := httpRoutes(systemNamespace, service, dests)
	if len(routes) == 0 {
		return nil
	}

	if clusterDomain.Status.ClusterDomain == "" {
		external = false
	}

	vs := newVirtualService(service)
	spec := v1alpha3.VirtualServiceSpec{
		Hosts:    []string{},
		Gateways: []string{privateGw},
		Http:     routes,
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

func newVirtualService(service *v1.Service) *v1alpha3.VirtualService {
	return constructors.NewVirtualService(service.Namespace, service.Name, v1alpha3.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
		},
	})
}

func isProtocolSupported(protocol v1.Protocol) bool {
	if protocol == v1.ProtocolHTTP || protocol == v1.ProtocolHTTP2 || protocol == v1.ProtocolGRPC || protocol == v1.ProtocolTCP {
		return true
	}
	return false
}
