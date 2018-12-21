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
	"github.com/rancher/rio/pkg/serviceset"
	"github.com/rancher/rio/pkg/settings"
	v1alpha3client "github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	privateGw              = "mesh"
	PublicDomainAnnotation = "rio.cattle.io/publicDomain"
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

func vsRoutes(publicPorts map[string]bool, service *v1.Service, dests []dest) ([]v1alpha3.HTTPRoute, bool) {
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
		ds := []dest{
			{
				host:   fmt.Sprintf("%s.rio-system.svc.cluster.local", fmt.Sprintf("cm-acme-http-solver-%d", adler32.Checksum([]byte(publicDomain)))),
				subset: "latest",
				weight: 100,
			},
		}
		_, route := newRoute(domains.GetPublicGateway(), true, pb, ds, false)
		route.Match[0].Uri = &v1alpha1.StringMatch{
			Prefix: "/.well-known/acme-challenge/",
		}
		route.Match[0].Authority = &v1alpha1.StringMatch{
			Prefix: publicDomain,
		}
		result = append(result, route)
	}

	containerlist.ForService(service)
	for _, con := range containerlist.ForService(service) {
		for _, exposed := range con.ExposedPorts {
			publicPort, route := newRoute(domains.GetPublicGateway(), false, &exposed.PortBinding, dests, true)
			if publicPort != "" {
				result = append(result, route)
			}
		}

		for _, binding := range con.PortBindings {
			publicPort, route := newRoute(domains.GetPublicGateway(), true, &binding, dests, true)
			if publicPort != "" {
				external = true
				publicPorts[publicPort] = true
				result = append(result, route)
			}
		}
	}

	return result, external
}

func newRoute(externalGW string, published bool, portBinding *v1.PortBinding, dests []dest, appendHttps bool) (string, v1alpha3.HTTPRoute) {
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

	for _, dest := range dests {
		route.Route = append(route.Route, v1alpha3.DestinationWeight{
			Destination: v1alpha3.Destination{
				Host:   dest.host,
				Subset: dest.subset,
				Port: v1alpha3.PortSelector{
					Number: uint32(portBinding.TargetPort),
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
	weight       int
}

func destsForService(name string, service *serviceset.ServiceSet) []dest {
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
			weight: weight,
			subset: rev.Spec.Revision.Version,
		})
	}

	result[0].weight = latestWeight
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

func vsFromService(stack *v1.Stack, name string, service *serviceset.ServiceSet) []runtime.Object {
	var result []runtime.Object

	serviceVS := vsFromSpec(stack, name, service.Service.Namespace, service.Service, destsForService(name, service)...)
	if serviceVS != nil {
		result = append(result, serviceVS)
	}

	for _, rev := range service.Revisions {
		revVs := vsFromSpec(stack, rev.Name, service.Service.Namespace, rev, dest{
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

func vsFromSpec(stack *v1.Stack, name, namespace string, service *v1.Service, dests ...dest) *v1alpha3client.VirtualService {
	publicPorts := map[string]bool{}

	routes, external := vsRoutes(publicPorts, service, dests)
	if len(routes) == 0 {
		return nil
	}

	vs := newVirtualService(stack, service)
	spec := v1alpha3.VirtualServiceSpec{
		Hosts:    []string{name},
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
