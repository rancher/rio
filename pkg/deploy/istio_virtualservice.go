package deploy

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"istio.io/api/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	PRIVATE_GW = "mesh"
)

func virtualservices(objs []runtime.Object, resources *StackResources) ([]runtime.Object, error) {
	var err error

	ns := ""
	for _, service := range resources.Services {
		objs, err = vsFromService(objs, service)
		if err != nil {
			return nil, err
		}
		if ns == "" {
			ns = service.Namespace
		}
	}

	return objs, nil
}

func coalescePort(port, targetPort int64) uint32 {
	if port <= 0 {
		return uint32(targetPort)
	}
	return uint32(port)
}

func vsRoutes(publicPorts map[uint32]bool, name, namespace string, serviceSpec *v1beta1.ServiceUnversionedSpec, dests []dest) ([]*v1alpha3.HTTPRoute, bool) {
	external := false
	var result []*v1alpha3.HTTPRoute

	for _, exposed := range serviceSpec.ExposedPorts {
		_, route := newRoute(getPublicGateway(name, namespace), false, &exposed.PortBinding, dests)
		if route != nil {
			result = append(result, route)
		}
	}

	for _, binding := range serviceSpec.PortBindings {
		publicPort, route := newRoute(getPublicGateway(name, namespace), true, &binding, dests)
		if route != nil {
			external = true
			publicPorts[publicPort] = true
			result = append(result, route)
		}
	}

	return result, external
}

func newRoute(externalGW string, published bool, portBinding *v1beta1.PortBinding, dests []dest) (uint32, *v1alpha3.HTTPRoute) {
	if portBinding.Protocol != "http" {
		return 0, nil
	}

	gw := []string{PRIVATE_GW}
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

func eachRev(revs map[string]v1beta1.ServiceRevision) []string {
	var keys []string
	for k := range revs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

type dest struct {
	host, subset string
	weight       int32
}

func destsForService(service *v1beta1.Service) []dest {
	latestWeight := 100
	result := []dest{
		{
			host:   service.Name,
			subset: "latest",
		},
	}

	for _, rev := range eachRev(service.Spec.Revisions) {
		revSpec := service.Spec.Revisions[rev]
		if latestWeight == 0 {
			// no more weight left
			continue
		}

		weight := min(revSpec.Weight, 100)
		if weight <= 0 {
			continue
		}

		weight = min(weight, latestWeight)
		latestWeight -= weight

		result = append(result, dest{
			host:   service.Name,
			weight: int32(weight),
		})
	}

	result[0].weight = int32(latestWeight)
	return result
}

func min(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func vsFromService(objs []runtime.Object, service *v1beta1.Service) ([]runtime.Object, error) {
	serviceVS := vsFromSpec(service.Name, "latest", service.Name, service.Namespace, &service.Spec.ServiceUnversionedSpec, destsForService(service)...)
	if serviceVS != nil {
		objs = append(objs, serviceVS)
	}

	for _, rev := range eachRev(service.Spec.Revisions) {
		mergeRevSpec, err := MergeRevisionToService(service, rev)
		if err != nil {
			return nil, err
		}

		revVs := vsFromSpec(service.Name, rev, service.Name+"-"+rev, service.Namespace, mergeRevSpec, dest{
			host:   service.Name,
			subset: rev,
			weight: 100,
		})
		if revVs != nil {
			objs = append(objs, revVs)
		}
	}

	return objs, nil
}

func vsFromSpec(serviceName, revision, name, namespace string, serviceSpec *v1beta1.ServiceUnversionedSpec, dests ...dest) *IstioObject {
	publicPorts := map[uint32]bool{}

	vs := newVirtualService(serviceName, revision, name, namespace)
	spec := &v1alpha3.VirtualService{
		Hosts:    []string{name},
		Gateways: []string{PRIVATE_GW},
	}
	vs.Spec = spec

	routes, external := vsRoutes(publicPorts, name, namespace, serviceSpec, dests)
	if len(routes) == 0 {
		return nil
	}

	spec.Http = routes

	if external && len(publicPorts) > 0 {
		externalGW := getPublicGateway(name, namespace)
		externalHost := getExternalDomain(name, namespace)
		spec.Hosts = append(spec.Hosts, externalHost)
		spec.Gateways = append(spec.Gateways, externalGW)

		var portList []string
		for p := range publicPorts {
			portList = append(portList, strconv.FormatUint(uint64(p), 10))
		}
		sort.Slice(portList, func(i, j int) bool {
			return portList[i] < portList[j]
		})

		vs.Annotations["rio.cattle.io/ports"] = strings.Join(portList, ",")
	}

	return vs
}

func getPublicGateway(name, namespace string) string {
	return fmt.Sprintf("external.%s.svc.cluster.local", settings.RioSystemNamespace)
}

func getExternalDomain(name, namespace string) string {
	return fmt.Sprintf("%s.%s.%s", name,
		strings.SplitN(namespace, "-", 2)[0], settings.ClusterDomain.Get())
}

func newVirtualService(serviceName, revision, name, namespace string) *IstioObject {
	return &IstioObject{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "VirtualService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: map[string]string{},
			Labels: map[string]string{
				"rio.cattle.io":           "true",
				"rio.cattle.io/service":   serviceName,
				"rio.cattle.io/revision":  revision,
				"rio.cattle.io/namespace": namespace,
			},
		},
	}
}
