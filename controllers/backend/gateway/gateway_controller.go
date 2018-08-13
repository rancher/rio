package gateway

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/rancher/norman/types/set"
	"github.com/rancher/rancher/pkg/controllers/user/approuter"
	"github.com/rancher/rancher/pkg/ticker"
	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3"
	"github.com/rancher/types/apis/apps/v1beta2"
	v12 "github.com/rancher/types/apis/core/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	all = "_all_"
)

var (
	addressTypes = []v1.NodeAddressType{
		v1.NodeExternalIP,
		v1.NodeInternalIP,
	}
)

type Controller struct {
	gateways                 v1alpha3.GatewayInterface
	gatewayLister            v1alpha3.GatewayLister
	virtualServiceLister     v1alpha3.VirtualServiceLister
	virtualServiceController v1alpha3.VirtualServiceController
	services                 v12.ServiceInterface
	serviceLister            v12.ServiceLister
	endpointsLister          v12.EndpointsLister
	nodeLister               v12.NodeLister
	deploymentsLister        v1beta2.DeploymentLister
	deployments              v1beta2.DeploymentInterface
	pods                     v12.PodInterface
	rdnsClient               *approuter.Client
	previousIPs              []string
}

func Register(ctx context.Context, rContext *types.Context) {
	rdnsClient := approuter.NewClient(rContext.Core.Secrets(""),
		rContext.Core.Secrets("").Controller().Lister(),
		settings.RioSystemNamespace)
	rdnsClient.SetBaseURL(settings.RDNSURL.Get())

	gc := &Controller{
		gateways:                 rContext.Networking.Gateways(""),
		gatewayLister:            rContext.Networking.Gateways("").Controller().Lister(),
		virtualServiceLister:     rContext.Networking.VirtualServices("").Controller().Lister(),
		virtualServiceController: rContext.Networking.VirtualServices("").Controller(),
		services:                 rContext.Core.Services(""),
		serviceLister:            rContext.Core.Services("").Controller().Lister(),
		endpointsLister:          rContext.Core.Endpoints("").Controller().Lister(),
		deploymentsLister:        rContext.Apps.Deployments("").Controller().Lister(),
		deployments:              rContext.Apps.Deployments(""),
		pods:                     rContext.Core.Pods(""),
		nodeLister:               rContext.Core.Nodes("").Controller().Lister(),
		rdnsClient:               rdnsClient,
	}
	rContext.Networking.VirtualServices("").Controller().AddHandler("gateway-controller", gc.sync)
	rContext.Core.Services("").Controller().AddHandler("gateway-service-controller", gc.serviceChanged)
	rContext.Core.Pods("").Controller().AddHandler("gateway-pod-controller", gc.podChanged)

	go func() {
		for range ticker.Context(ctx, 6*time.Hour) {
			gc.renew()
		}
	}()
}

func (g *Controller) podChanged(key string, pod *v1.Pod) error {
	if pod == nil {
		return nil
	}

	if pod.Labels["gateway"] != "external" {
		return nil
	}

	if pod.Status.HostIP == "" {
		return nil
	}

	hostIP := pod.Annotations["rio.cattle.io/host-ip"]
	if hostIP != pod.Status.HostIP {
		g.virtualServiceController.Enqueue("", all)

		pod = pod.DeepCopy()
		if pod.Annotations == nil {
			pod.Annotations = map[string]string{}
		}

		pod.Annotations["rio.cattle.io/host-ip"] = pod.Status.HostIP
		_, err := g.pods.Update(pod)
		return err
	}

	return nil
}

func (g *Controller) serviceChanged(key string, service *v1.Service) error {
	if service != nil && service.Name == settings.IstionExternalLB {
		g.virtualServiceController.Enqueue("", all)
	}
	return nil
}

func (g *Controller) sync(key string, service *v1alpha3.VirtualService) error {
	if key != all {
		g.virtualServiceController.Enqueue("", all)
		return nil
	}

	vss, err := g.virtualServiceLister.List("", labels.Everything())
	if err != nil {
		return err
	}

	ports := map[string]bool{}
	for _, vs := range vss {
		for _, port := range getPorts(vs) {
			ports[port] = true
		}
	}

	ns := namespace.StackNamespace(settings.RioSystemNamespace, settings.IstioStackName.Get())

	ips, hostPorts, err := g.setServicePorts(ns, ports)
	if err != nil {
		return err
	}

	if err := g.updateDomain(ips); err != nil {
		return err
	}

	if err := g.setGatewayPorts(ns, ports); err != nil {
		return err
	}

	return g.setHostPorts(ns, hostPorts, ports)
}

func getPorts(service *v1alpha3.VirtualService) []string {
	ports, ok := service.Annotations["rio.cattle.io/ports"]
	if !ok || ports == "" {
		return nil
	}

	return strings.Split(ports, ",")
}

func (g *Controller) setGatewayPorts(ns string, ports map[string]bool) error {
	existingPorts := map[string]bool{}

	gw, err := g.gatewayLister.Get(ns, settings.IstionExternalGateway)
	if err != nil {
		return err
	}

	for _, server := range gw.Spec.Servers {
		existingPorts[strconv.FormatUint(uint64(server.Port.Number), 10)] = true
	}

	if !set.Changed(ports, existingPorts) {
		return nil
	}

	gw = gw.DeepCopy()
	gw.Spec.Servers = nil

	for portStr := range ports {
		port, err := strconv.ParseUint(portStr, 10, 32)
		if err != nil {
			continue
		}

		gw.Spec.Servers = append(gw.Spec.Servers, &v1alpha3.Server{
			Hosts: []string{
				"*",
			},
			Port: &v1alpha3.Port{
				Protocol: "http",
				Number:   uint32(port),
			},
		})
	}

	_, err = g.gateways.Update(gw)
	return err
}

func getNodeIP(node *v1.Node) string {
	for _, addrType := range addressTypes {
		for _, addr := range node.Status.Addresses {
			if addrType == addr.Type {
				return addr.Address
			}
		}
	}

	return ""
}

func (g *Controller) setServicePorts(ns string, ports map[string]bool) ([]string, bool, error) {
	existingPorts := map[string]bool{}

	svc, err := g.serviceLister.Get(ns, settings.IstionExternalLB)
	if err != nil {
		return nil, false, err
	}

	for _, port := range svc.Spec.Ports {
		existingPorts[strconv.FormatUint(uint64(port.Port), 10)] = true
	}

	var ips []string
	hostPorts := false
	for _, ingress := range svc.Status.LoadBalancer.Ingress {
		if ingress.Hostname == "localhost" {
			ips = append(ips, "127.0.0.1")
		} else if ingress.IP != "" {
			ips = append(ips, ingress.IP)
		}
	}

	if len(ips) == 0 {
		hostPorts = true
		ep, err := g.endpointsLister.Get(svc.Namespace, svc.Name)
		if err != nil {
			return nil, false, err
		}

		for _, subset := range ep.Subsets {
			for _, addr := range subset.Addresses {
				if addr.NodeName == nil {
					continue
				}

				node, err := g.nodeLister.Get("", *addr.NodeName)
				if err != nil {
					return nil, false, err
				}

				nodeIP := getNodeIP(node)
				if nodeIP != "" {
					ips = append(ips, nodeIP)
				}
			}
		}
	}

	// you can't update a service to zero ports
	if !set.Changed(ports, existingPorts) || len(ports) == 0 {
		return ips, hostPorts, nil
	}

	svc = svc.DeepCopy()
	svc.Spec.Ports = nil

	for portStr := range ports {
		port, err := strconv.ParseInt(portStr, 10, 32)
		if err != nil {
			continue
		}

		svc.Spec.Ports = append(svc.Spec.Ports, v1.ServicePort{
			Name:       "http-" + portStr,
			Protocol:   v1.ProtocolTCP,
			Port:       int32(port),
			TargetPort: intstr.FromInt(int(port)),
		})
	}

	_, err = g.services.Update(svc)
	return ips, hostPorts, err
}

func (g *Controller) setHostPorts(ns string, hostPorts bool, ports map[string]bool) error {
	gatewayDep, err := g.deploymentsLister.Get(ns, settings.IstionExternalGatewayDeployment)
	if err != nil {
		return err
	}

	existingPorts := map[string]bool{}
	for _, port := range gatewayDep.Spec.Template.Spec.Containers[0].Ports {
		if port.HostPort > 0 {
			existingPorts[strconv.FormatInt(int64(port.HostPort), 10)] = true
		}
	}

	if !hostPorts {
		ports = nil
	}

	toCreate, toDelete, _ := set.Diff(ports, existingPorts)
	if len(toCreate) == 0 && len(toDelete) == 0 {
		return nil
	}

	if hostPorts && len(toCreate) == 0 {
		// For host ports we don't care too much about closing ports.  So if all we are doing is deleting ports
		// then just skip it for now
		return nil
	}

	gatewayDep = gatewayDep.DeepCopy()
	gatewayDep.Spec.Template.Spec.Containers[0].Ports = nil
	for portStr := range ports {
		p, err := strconv.ParseInt(portStr, 10, 0)
		if err != nil {
			return err
		}

		gatewayDep.Spec.Template.Spec.Containers[0].Ports =
			append(gatewayDep.Spec.Template.Spec.Containers[0].Ports, v1.ContainerPort{
				Name:          "port-" + portStr,
				HostPort:      int32(p),
				Protocol:      v1.ProtocolTCP,
				ContainerPort: int32(p),
			})
	}
	_, err = g.deployments.Update(gatewayDep)
	return err
}
