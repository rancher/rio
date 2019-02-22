package populate

import (
	"fmt"
	"hash/adler32"
	"sort"
	"strconv"
	"strings"

	"github.com/knative/pkg/apis/istio/common/v1alpha1"
	"github.com/knative/pkg/apis/istio/v1alpha3"
	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/api/service"
	"github.com/rancher/rio/features/routing/pkg/domains"
	"github.com/rancher/rio/features/stack/controllers/service/populate/containerlist"
	"github.com/rancher/rio/features/stack/controllers/service/populate/servicelabels"
	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/pkg/serviceset"
	"github.com/rancher/rio/pkg/settings"
	v1alpha3client "github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3"
	v1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	privateGw              = "mesh"
	PublicDomainAnnotation = "rio.cattle.io/publicDomain"
	RioNameHeader          = "X-Rio-ServiceName"
	RioNamespaceHeader     = "X-Rio-Namespace"
	RioPortHeader          = "X-Rio-ServicePort"
)

func virtualServices(stack *v1.Stack, services []*v1.Service, service *v1.Service, os *objectset.ObjectSet) error {
	serviceSets, err := serviceset.CollectionServices(services)
	if err != nil {
		return err
	}

	serviceSet, ok := serviceSets[service.Name]
	if !ok {
		return nil
	}

	os.Add(vsFromService(stack, service.Name, serviceSet)...)

	return nil
}

func vsRoutes(publicPorts map[string]bool, service *v1.Service, dests []Dest) ([]v1alpha3.HTTPRoute, bool) {
	external := false
	var result []v1alpha3.HTTPRoute

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
		ds := []Dest{
			{
				Host:   fmt.Sprintf("%s.rio-system.svc.cluster.local", fmt.Sprintf("cm-acme-http-solver-%d", adler32.Checksum([]byte(publicDomain)))),
				Subset: "latest",
				Weight: 100,
			},
		}
		_, route := newRoute(domains.GetPublicGateway(), true, pb, ds, false, false, nil)
		route.Match[0].Uri = &v1alpha1.StringMatch{
			Prefix: "/.well-known/acme-challenge/",
		}
		route.Match[0].Authority = &v1alpha1.StringMatch{
			Prefix: publicDomain,
		}
		result = append(result, route)
	}

	containerlist.ForService(service)
	enableAutoScale := service.Spec.AutoScale != nil
	for _, con := range containerlist.ForService(service) {
		for _, exposed := range con.ExposedPorts {
			publicPort, route := newRoute(domains.GetPublicGateway(), false, &exposed.PortBinding, dests, true, enableAutoScale, service)
			if publicPort != "" {
				result = append(result, route)
			}
		}

		for _, binding := range con.PortBindings {
			publicPort, route := newRoute(domains.GetPublicGateway(), true, &binding, dests, true, enableAutoScale, service)
			if publicPort != "" {
				external = true
				publicPorts[publicPort] = true
				result = append(result, route)
			}
		}
	}

	return result, external
}

func newRoute(externalGW string, published bool, portBinding *v1.PortBinding, dests []Dest, appendHttps bool, autoscale bool, svc *v1.Service) (string, v1alpha3.HTTPRoute) {
	route := v1alpha3.HTTPRoute{}

	if _, ok := service.SupportedProtocol[portBinding.Protocol]; !ok {
		return "", route
	}

	gw := []string{privateGw}
	if published {
		gw = append(gw, externalGW)
	}

	httpPort, _ := strconv.ParseUint(settings.DefaultHTTPOpenPort.Get(), 10, 64)
	httpsPort, _ := strconv.ParseUint(settings.DefaultHTTPSOpenPort.Get(), 10, 64)
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
					Host: fmt.Sprintf("%s.%s.svc.cluster.local", namespace.HashIfNeed("gateway", settings.AutoScaleStack, settings.RioSystemNamespace), settings.CloudNamespace),
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

func DestsForService(name, stackName, projectName string, service *serviceset.ServiceSet) []Dest {
	latestWeight := 100
	svcName := fmt.Sprintf("%s-%s", name, namespace.StackNamespaceOnlyHash(projectName, stackName))
	result := []Dest{
		{
			Host:   fmt.Sprintf("%s.%s.svc.cluster.local", svcName, settings.CloudNamespace),
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

func vsFromService(stack *v1.Stack, name string, service *serviceset.ServiceSet) []runtime.Object {
	var result []runtime.Object

	serviceVS := VsFromSpec(stack, name, service.Service.Namespace, service.Service, DestsForService(name, stack.Name, stack.Namespace, service)...)
	if serviceVS != nil {
		result = append(result, serviceVS)
	}

	for _, rev := range service.Revisions {
		revVs := VsFromSpec(stack, rev.Name, service.Service.Namespace, rev, Dest{
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

func VsFromSpec(stack *v1.Stack, name, namespace string, service *v1.Service, dests ...Dest) *v1alpha3client.VirtualService {
	publicPorts := map[string]bool{}

	routes, external := vsRoutes(publicPorts, service, dests)
	if len(routes) == 0 {
		return nil
	}

	vs := newVirtualService(stack, service)
	spec := v1alpha3.VirtualServiceSpec{
		Hosts:    []string{},
		Gateways: []string{privateGw},
		Http:     routes,
	}

	if external && len(publicPorts) > 0 {
		externalGW := domains.GetPublicGateway()
		externalHost := domains.GetExternalDomain(name, namespace, stack.Namespace)
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
	vs.Spec = spec

	return vs
}

func newVirtualService(stack *v1.Stack, service *v1.Service) *v1alpha3client.VirtualService {
	return v1alpha3client.NewVirtualService(service.Namespace, service.Name, v1alpha3client.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
			Labels:      servicelabels.RioOnlyServiceLabels(stack, service),
		},
	})
}
