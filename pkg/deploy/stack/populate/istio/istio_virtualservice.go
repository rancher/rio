package istio

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/rancher/rio/pkg/deploy/stack/input"
	"github.com/rancher/rio/pkg/deploy/stack/output"
	"github.com/rancher/rio/pkg/deploy/stack/populate/containerlist"
	"github.com/rancher/rio/pkg/deploy/stack/populate/service"
	"github.com/rancher/rio/pkg/deploy/stack/populate/servicelabels"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
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

	return result, nil
}

func coalescePort(port, targetPort int64) uint32 {
	if port <= 0 {
		return uint32(targetPort)
	}
	return uint32(port)
}

func vsRoutes(publicPorts map[uint32]bool, name, namespace string, service *v1beta1.Service, dests []dest) ([]*v1alpha3.HTTPRoute, bool) {
	external := false
	var result []*v1alpha3.HTTPRoute

	containerlist.ForService(service)
	for _, con := range containerlist.ForService(service) {
		for _, exposed := range con.ExposedPorts {
			_, route := newRoute(getPublicGateway(), false, &exposed.PortBinding, dests)
			if route != nil {
				result = append(result, route)
			}
		}

		for _, binding := range con.PortBindings {
			publicPort, route := newRoute(getPublicGateway(), true, &binding, dests)
			if route != nil {
				external = true
				publicPorts[publicPort] = true
				result = append(result, route)
			}
		}
	}

	return result, external
}

func newRoute(externalGW string, published bool, portBinding *v1beta1.PortBinding, dests []dest) (uint32, *v1alpha3.HTTPRoute) {
	if portBinding.Protocol != "http" {
		return 0, nil
	}

	gw := []string{privateGw}
	if published {
		gw = append(gw, externalGW)
	}

	sourcePort := coalescePort(portBinding.Port, portBinding.TargetPort)
	if sourcePort <= 0 {
		return 0, nil
	}

	route := &v1alpha3.HTTPRoute{
		Match: []*v1alpha3.HTTPMatchRequest{
			{
				Port:     sourcePort,
				Gateways: gw,
			},
		},
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

	return sourcePort, route
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
			host:   name,
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
			host:   name,
			subset: rev.Spec.Revision.Version,
			weight: 100,
		})
		if revVs != nil {
			result = append(result, revVs)
		}
	}

	return result
}

func vsFromSpec(stack *input.Stack, name, namespace string, service *v1beta1.Service, dests ...dest) *output.IstioObject {
	publicPorts := map[uint32]bool{}

	routes, external := vsRoutes(publicPorts, name, namespace, service, dests)
	if len(routes) == 0 {
		return nil
	}

	vs := newVirtualService(stack, service)
	spec := &v1alpha3.VirtualService{
		Hosts:    appendStringWithPort(nil, name, publicPorts),
		Gateways: []string{privateGw},
		Http:     routes,
	}
	vs.Spec = spec

	if external && len(publicPorts) > 0 {
		externalGW := getPublicGateway()
		externalHost := getExternalDomain(name, namespace)
		spec.Gateways = append(spec.Gateways, externalGW)
		spec.Hosts = appendStringWithPort(spec.Hosts, externalHost, publicPorts)

		var portList []string
		for p := range publicPorts {
			portList = append(portList, strconv.FormatUint(uint64(p), 10))
		}
		sort.Slice(portList, func(i, j int) bool {
			return portList[i] < portList[j]
		})

		vs.Annotations["rio.cattle.io/ports"] = strings.Join(portList, ",")
	}

	if service.Annotations[PublicDomainAnnotation] != "" {
		spec.Hosts = append(spec.Hosts, service.Annotations[PublicDomainAnnotation])
	}

	return vs
}

func appendStringWithPort(base []string, host string, ports map[uint32]bool) []string {
	for port := range ports {
		if port == 80 || port == 443 {
			base = append(base, host)
		} else {
			base = append(base, fmt.Sprintf("%s:%d", host, port))
		}
	}

	return base
}

func getPublicGateway() string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", settings.IstioGateway, settings.RioSystemNamespace)
}

func getExternalDomain(name, namespace string) string {
	return fmt.Sprintf("%s.%s.%s", name,
		strings.SplitN(namespace, "-", 2)[0], settings.ClusterDomain.Get())
}

func newVirtualService(stack *input.Stack, service *v1beta1.Service) *output.IstioObject {
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
