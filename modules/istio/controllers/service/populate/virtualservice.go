package populate

import (
	"fmt"
	"hash/adler32"
	"strconv"

	"github.com/knative/pkg/apis/istio/common/v1alpha1"
	"github.com/knative/pkg/apis/istio/v1alpha3"
	"github.com/rancher/rio/modules/istio/pkg/domains"
	projectv1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/pkg/services"
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

func virtualServices(namespace string, clusterDomain *projectv1.ClusterDomain, serviceSet *serviceset.ServiceSet, service *v1.Service, os *objectset.ObjectSet) error {
	os.Add(virtualServiceFromService(namespace, clusterDomain, serviceSet)...)
	return nil
}

func httpRoutes(systemNamespace string, service *v1.Service, dests []Dest) ([]v1alpha3.HTTPRoute, bool) {
	external := false
	var result []v1alpha3.HTTPRoute

	enableAutoScale := service.Spec.AutoscaleConfig.MinScale != service.Spec.AutoscaleConfig.MaxScale
	for _, port := range service.Spec.Ports {
		publicPort, route := newRoute(domains.GetPublicGateway(systemNamespace), !port.InternalOnly, port, dests, true, enableAutoScale, service)
		if publicPort != "" {
			external = true
			result = append(result, route)
		}
	}

	return result, external
}

func newRoute(externalGW string, published bool, portBinding v1.ContainerPort, dests []Dest, appendHttps bool, autoscale bool, svc *v1.Service) (string, v1alpha3.HTTPRoute) {
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
		route.AppendHeaders[RioPortHeader] = strconv.Itoa(int(portBinding.TargetPort))
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
						Number: uint32(portBinding.TargetPort),
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
	var result []Dest
	for _, rev := range service.Revisions {
		_, ver := services.AppAndVersion(rev)
		weight := rev.Spec.ServiceRevision.Weight
		if rev.Status.WeightOverride != nil {
			weight = *rev.Status.WeightOverride
		}
		result = append(result, Dest{
			Host:   fmt.Sprintf("%s.%s.svc.cluster.local", name, namespace),
			Weight: weight,
			Subset: ver,
		})
	}

	return result
}

func virtualServiceFromService(namespace string, clusterDomain *projectv1.ClusterDomain, service *serviceset.ServiceSet) []runtime.Object {
	var result []runtime.Object

	for _, rev := range service.Revisions {
		_, version := services.AppAndVersion(rev)
		weight := rev.Spec.ServiceRevision.Weight
		if rev.Status.WeightOverride != nil {
			weight = *rev.Status.WeightOverride
		}
		revVs := VirtualServiceFromSpec(namespace, rev.Name, rev.Namespace, clusterDomain, rev, Dest{
			Host:   rev.Name,
			Subset: version,
			Weight: weight,
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
		Hosts:    []string{service.Name},
		Gateways: []string{privateGw},
		Http:     routes,
	}

	acmeSolverBinding := v1.ContainerPort{
		Port:       80,
		TargetPort: 8089,
		Protocol:   v1.ProtocolTCP,
	}
	for _, publicDomain := range service.Status.PublicDomains {
		if publicDomain == "" {
			continue
		}
		spec.Hosts = append(spec.Hosts, publicDomain)
		ds := []Dest{
			{
				Host:   fmt.Sprintf("%s.%s.svc.cluster.local", fmt.Sprintf("cm-acme-http-solver-%d", adler32.Checksum([]byte(publicDomain))), systemNamespace),
				Subset: "latest",
				Weight: 100,
			},
		}
		_, route := newRoute(domains.GetPublicGateway(systemNamespace), true, acmeSolverBinding, ds, false, false, nil)
		route.Match[0].Uri = &v1alpha1.StringMatch{
			Prefix: "/.well-known/acme-challenge/",
		}
		route.Match[0].Authority = &v1alpha1.StringMatch{
			Prefix: publicDomain,
		}
		spec.Http = append(spec.Http, route)
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
