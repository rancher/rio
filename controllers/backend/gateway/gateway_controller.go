package gateway

import (
	"strconv"
	"sync"
	"time"

	"strings"

	"context"

	"github.com/rancher/norman/types/set"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3"
	v1beta22 "github.com/rancher/types/apis/apps/v1beta2"
	v12 "github.com/rancher/types/apis/core/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	refreshInterval = 5 * time.Minute
)

type GatewayController struct {
	sync.Mutex

	ports                map[string]bool
	gatewayLister        v1alpha3.GatewayLister
	virtualServiceLister v1alpha3.VirtualServiceLister
	gateways             v1alpha3.GatewayInterface
	daemonSets           v1beta22.DaemonSetInterface
	nodeLister           v12.NodeLister
	services             v12.ServiceInterface
	lastRefresh          time.Time
}

func Register(ctx context.Context, rContext *types.Context) {
	gc := &GatewayController{
		ports:                map[string]bool{},
		gatewayLister:        rContext.Networking.Gateways("").Controller().Lister(),
		virtualServiceLister: rContext.Networking.VirtualServices("").Controller().Lister(),
		gateways:             rContext.Networking.Gateways(""),
		daemonSets:           rContext.Apps.DaemonSets(""),
		services:             rContext.Core.Services(""),
		nodeLister:           rContext.Core.Nodes("").Controller().Lister(),
	}
	rContext.Networking.VirtualServices("").Controller().AddHandler("gateway-controller", gc.sync)
}

func (g *GatewayController) sync(key string, service *v1alpha3.VirtualService) error {
	if service == nil {
		return nil
	}

	g.Lock()
	if time.Now().Sub(g.lastRefresh) > refreshInterval {
		g.refresh()
	}
	g.Unlock()

	return g.addPorts(getPorts(service)...)
}

func getPorts(service *v1alpha3.VirtualService) []string {
	ports, ok := service.Annotations["rio.cattle.io/ports"]
	if !ok || ports == "" {
		return nil
	}

	return strings.Split(ports, ",")
}

func (g *GatewayController) refresh() error {
	now := time.Now()
	existingPorts := map[string]bool{}
	newPorts := map[string]bool{}

	gw, err := g.gatewayLister.Get(settings.RioSystemNamespace, settings.IstionExternalGateway)
	if err == nil {
		for _, server := range gw.Spec.Servers {
			existingPorts[strconv.FormatUint(uint64(server.Port.Number), 10)] = true
		}
	}

	vss, err := g.virtualServiceLister.List("", labels.Everything())
	if err != nil {
		return err
	}

	for _, vs := range vss {
		newPorts, _ = addPorts(newPorts, getPorts(vs)...)
	}

	toCreate, toDelete, _ := set.Diff(newPorts, existingPorts)
	if len(toCreate) > 0 || len(toDelete) > 0 {
		err = g.createGateway(newPorts)
	}

	if err != nil {
		return err
	}

	g.lastRefresh = now
	g.ports = newPorts
	return nil
}

func (g *GatewayController) addPorts(ports ...string) error {
	g.Lock()
	defer g.Unlock()

	newPorts, add := addPorts(g.ports, ports...)
	if !add {
		return nil
	}

	return g.createGateway(newPorts)
}

func (g *GatewayController) createGateway(newPorts map[string]bool) error {
	if err := g.deployDummy(newPorts); err != nil {
		return err
	}

	spec := v1alpha3.GatewaySpec{
		Selector: map[string]string{
			"gateway": "external",
		},
	}

	for portStr := range newPorts {
		port, err := strconv.ParseUint(portStr, 10, 32)
		if err != nil {
			continue
		}

		spec.Servers = append(spec.Servers, &v1alpha3.Server{
			Hosts: []string{
				"*",
			},
			Port: &v1alpha3.Port{
				Protocol: "http",
				Number:   uint32(port),
			},
		})
	}

	gw, err := g.gatewayLister.Get(settings.RioSystemNamespace, settings.IstionExternalGateway)
	if errors.IsNotFound(err) {
		if len(spec.Servers) == 0 {
			return nil
		}
		_, err := g.gateways.Create(&v1alpha3.Gateway{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Gateway",
				APIVersion: "networking.istio.io/v1alpha3",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: settings.RioSystemNamespace,
				Name:      settings.IstionExternalGateway,
			},
			Spec: spec,
		})
		return err
	} else if err != nil {
		return err
	}

	if len(spec.Servers) == 0 {
		err = g.gateways.DeleteNamespaced(gw.Namespace, gw.Name, nil)
	} else {
		gw.Spec = spec
		_, err = g.gateways.Update(gw)
	}

	if err != nil {
		return err
	}

	g.ports = newPorts

	return err
}

func (g *GatewayController) getD4xIP() string {
	node, err := g.nodeLister.Get("", "docker-for-desktop")
	if err != nil {
		return ""
	}

	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeInternalIP {
			return addr.Address
		}
	}

	return ""
}

func (g *GatewayController) deployDummy(ports map[string]bool) error {
	ip := g.getD4xIP()
	if ip == "" {
		return nil
	}

	ds := newDummyDS(ip)
	existing, err := g.daemonSets.GetNamespaced(ds.Namespace, ds.Name, metav1.GetOptions{})
	if err == nil {
		existing.Spec = ds.Spec
		_, err = g.daemonSets.Update(existing)
	} else {
		_, err = g.daemonSets.Create(ds)
	}

	if err != nil {
		return err
	}

	lb := newService()
	for p := range ports {
		portNum, err := strconv.Atoi(p)
		if err != nil {
			logrus.Errorf("failed to parse port %s: %v", p, err)
			continue
		}
		lb.Spec.Ports = append(lb.Spec.Ports, v1.ServicePort{
			TargetPort: intstr.FromInt(portNum),
			Port:       int32(portNum),
			Name:       "port-" + p,
			Protocol:   v1.ProtocolTCP,
		})
	}

	existingLB, err := g.services.GetNamespaced(ds.Namespace, ds.Name, metav1.GetOptions{})
	if err == nil {
		lb.Spec.ClusterIP = existingLB.Spec.ClusterIP
		existingLB.Spec = lb.Spec
		_, err := g.services.Update(existingLB)
		return err
	} else {
		_, err := g.services.Create(lb)
		return err
	}
}

func addPorts(existingPorts map[string]bool, ports ...string) (map[string]bool, bool) {
	newPorts := map[string]bool{}
	add := false

	for _, port := range ports {
		if _, ok := existingPorts[port]; ok {
			continue
		}
		add = true
		newPorts[port] = true
	}

	if !add {
		return nil, false
	}

	for k, v := range existingPorts {
		newPorts[k] = v
	}

	return newPorts, true
}

func newService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"rio.cattle.io": "true",
			},
			Namespace: settings.RioSystemNamespace,
			Name:      "d4x",
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"rio.cattle.io": "true",
				"app":           "d4x",
			},
			Type: v1.ServiceTypeLoadBalancer,
		},
	}
}

func newDummyDS(ip string) *v1beta2.DaemonSet {
	return &v1beta2.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1beta2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"rio.cattle.io": "true",
			},
			Namespace: settings.RioSystemNamespace,
			Name:      "d4x",
		},
		Spec: v1beta2.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"rio.cattle.io": "true",
					"app":           "d4x",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"rio.cattle.io": "true",
						"app":           "d4x",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "d4x",
							Image: settings.RioFullImage(),
							Command: []string{
								"/local-proxy",
							},
							Env: []v1.EnvVar{
								{
									Name:  "TARGET",
									Value: ip,
								},
							},
							SecurityContext: &v1.SecurityContext{
								Capabilities: &v1.Capabilities{
									Add: []v1.Capability{
										"NET_ADMIN",
									},
								},
							},
						},
					},
					Affinity: &v1.Affinity{
						NodeAffinity: &v1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
								NodeSelectorTerms: []v1.NodeSelectorTerm{
									{
										MatchExpressions: []v1.NodeSelectorRequirement{
											{
												Key:      "kubernetes.io/hostname",
												Operator: v1.NodeSelectorOpIn,
												Values: []string{
													"docker-for-desktop",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
