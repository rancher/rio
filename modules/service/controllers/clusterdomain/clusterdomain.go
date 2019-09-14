package clusterdomain

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/rancher/rio/modules/istio/pkg/domains"
	"github.com/rancher/rio/modules/istio/pkg/parse"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	adminv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	services2 "github.com/rancher/rio/pkg/services"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/generic"
	"github.com/rancher/wrangler/pkg/objectset"
	"github.com/rancher/wrangler/pkg/relatedresource"
	"github.com/rancher/wrangler/pkg/trigger"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	gatewayIngress = "gateway-ingress"

	nodeSelectorLabel   = "rio.cattle.io/gateway"
	serviceDomainUpdate = "service-domain-update"
	appDomainHandler    = "app-domain-update"
	routerDomainUpdate  = "router-domain-updater"
)

var (
	addressTypes = []v1.NodeAddressType{
		v1.NodeExternalIP,
		v1.NodeInternalIP,
	}

	evalTrigger trigger.Trigger
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		apply: rContext.Apply.WithSetID(gatewayIngress).WithStrictCaching().
			WithCacheTypes(rContext.K8sNetworking.Networking().V1beta1().Ingress()),
		namespace:         rContext.Namespace,
		apps:              rContext.Rio.Rio().V1().App(),
		services:          rContext.Rio.Rio().V1().Service(),
		routers:           rContext.Rio.Rio().V1().Router().Cache(),
		publicDomainCache: rContext.Global.Admin().V1().PublicDomain().Cache(),
		clusterDomain:     rContext.Global.Admin().V1().ClusterDomain(),
		secretCache:       rContext.Core.Core().V1().Secret().Cache(),
		nodeCache:         rContext.Core.Core().V1().Node().Cache(),
		endpointCache:     rContext.Core.Core().V1().Endpoints().Cache(),
	}

	relatedresource.Watch(ctx, "app-clusterdomain", h.resolveApp,
		rContext.Rio.Rio().V1().App(),
		rContext.Global.Admin().V1().ClusterDomain())

	relatedresource.Watch(ctx, "publicdomain-clusterdomain", h.resolve,
		rContext.Rio.Rio().V1().Service(),
		rContext.Global.Admin().V1().ClusterDomain())

	relatedresource.Watch(ctx, "router-clusterdomain", h.resolveRouter,
		rContext.Rio.Rio().V1().Router(),
		rContext.Global.Admin().V1().ClusterDomain())

	relatedresource.Watch(ctx, "cluster-domain-service", h.resolve,
		rContext.Global.Admin().V1().ClusterDomain(),
		rContext.Global.Admin().V1().PublicDomain())

	if constants.UseIPAddress == "" {
		switch constants.InstallMode {
		case constants.InstallModeSvclb:
			rContext.Core.Core().V1().Service().OnChange(ctx, "endpoints-serviceloadbalancer", h.syncServiceLoadbalancer)
		case constants.InstallModeHostport:
			rContext.Core.Core().V1().Endpoints().OnChange(ctx, "endpoints", h.syncEndpoint)
		case constants.InstallModeIngress:
			rContext.K8sNetworking.Networking().V1beta1().Ingress().OnChange(ctx, "ingress-endpoints", h.syncIngress)
		}
	} else {
		addresses := strings.Split(constants.UseIPAddress, ",")
		if err := h.updateClusterDomain(addresses); err != nil {
			return err
		}
	}

	rContext.Rio.Rio().V1().Service().OnChange(ctx, serviceDomainUpdate, riov1controller.UpdateServiceOnChange(rContext.Rio.Rio().V1().Service().Updater(), h.syncDomain))
	rContext.Rio.Rio().V1().App().OnChange(ctx, appDomainHandler, riov1controller.UpdateAppOnChange(rContext.Rio.Rio().V1().App().Updater(), h.syncAppDomain))
	rContext.Rio.Rio().V1().Router().AddGenericHandler(ctx, routerDomainUpdate, generic.UpdateOnChange(rContext.Rio.Rio().V1().Router().Updater(), h.syncRouterDomain))

	rContext.Global.Admin().V1().ClusterDomain().OnChange(ctx, "cluster-domain-gateway-ingress", h.syncClusterIngress)
	return nil
}

type handler struct {
	namespace         string
	apply             apply.Apply
	serviceApply      apply.Apply
	apps              riov1controller.AppController
	services          riov1controller.ServiceController
	routers           riov1controller.RouterCache
	publicDomainCache adminv1controller.PublicDomainCache
	clusterDomain     adminv1controller.ClusterDomainController
	secretCache       corev1controller.SecretCache
	nodeCache         corev1controller.NodeCache
	endpointCache     corev1controller.EndpointsCache
}

func (h handler) syncClusterIngress(key string, obj *adminv1.ClusterDomain) (*adminv1.ClusterDomain, error) {
	if obj == nil || obj.DeletionTimestamp != nil || obj.Name != constants.ClusterDomainName {
		return obj, nil
	}

	os := objectset.NewObjectSet()
	domain := ""
	if obj.Status.ClusterDomain != "" {
		domain = fmt.Sprintf("*.%s", obj.Status.ClusterDomain)
	}

	if constants.InstallMode == constants.InstallModeIngress {
		ingress := constructors.NewIngress(h.namespace, constants.ClusterIngressName, networkingv1beta1.Ingress{
			Spec: networkingv1beta1.IngressSpec{
				Rules: []networkingv1beta1.IngressRule{
					{
						Host: domain,
						IngressRuleValue: networkingv1beta1.IngressRuleValue{
							HTTP: &networkingv1beta1.HTTPIngressRuleValue{
								Paths: []networkingv1beta1.HTTPIngressPath{
									{
										Path: "/rio-gateway",
										Backend: networkingv1beta1.IngressBackend{
											ServiceName: constants.GatewayName,
											ServicePort: intstr.FromInt(80),
										},
									},
								},
							},
						},
					},
				},
			},
		})
		os.Add(ingress)
	}

	return obj, h.apply.Apply(os)
}

func (h handler) updateClusterDomain(addresses []string) error {
	clusterDomain, err := h.clusterDomain.Cache().Get(h.namespace, constants.ClusterDomainName)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if clusterDomain == nil {
		return err
	}

	deepcopy := clusterDomain.DeepCopy()
	var address []adminv1.Address
	for _, ip := range addresses {
		address = append(address, adminv1.Address{IP: ip})
	}
	if !reflect.DeepEqual(deepcopy.Spec.Addresses, address) {
		logrus.Infof("Updating cluster domain to address %v", addresses)
	}
	deepcopy.Spec.Addresses = address

	if _, err := h.clusterDomain.Update(deepcopy); err != nil {
		return err
	}

	return err
}

func (h handler) syncServiceLoadbalancer(key string, obj *v1.Service) (*v1.Service, error) {
	if obj == nil {
		return obj, nil
	}

	if obj.Spec.Selector["app"] != constants.GatewayName || obj.Namespace != h.namespace || obj.Spec.Type != v1.ServiceTypeLoadBalancer {
		return obj, nil
	}

	var address []string
	for _, ingress := range obj.Status.LoadBalancer.Ingress {
		if ingress.Hostname == "localhost" {
			ingress.IP = "127.0.0.1"
		}
		address = append(address, ingress.IP)
	}

	if err := h.updateClusterDomain(address); err != nil {
		return obj, err
	}
	return obj, nil
}

func (h handler) syncEndpoint(key string, endpoint *v1.Endpoints) (*v1.Endpoints, error) {
	if endpoint == nil {
		return nil, nil
	}
	if endpoint.Namespace != h.namespace || endpoint.Name != constants.GatewayName {
		return endpoint, nil
	}

	var ips []string
	for _, subset := range endpoint.Subsets {
		for _, addr := range subset.Addresses {
			if addr.NodeName == nil {
				continue
			}

			node, err := h.nodeCache.Get(*addr.NodeName)
			if err != nil {
				return nil, err
			}

			nodeIP := getNodeIP(node)
			if nodeIP != "" {
				ips = append(ips, nodeIP)
			}
		}
	}
	if err := h.updateClusterDomain(ips); err != nil {
		return endpoint, err
	}
	return endpoint, nil
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

func (h handler) syncIngress(key string, ingress *networkingv1beta1.Ingress) (*networkingv1beta1.Ingress, error) {
	if ingress == nil {
		return ingress, nil
	}

	if ingress.Namespace == h.namespace && ingress.Name == constants.ClusterIngressName {
		var ips []string
		for _, ip := range ingress.Status.LoadBalancer.Ingress {
			if ip.IP != "" {
				ips = append(ips, ip.IP)
			}
		}
		return ingress, h.updateClusterDomain(ips)
	}
	return ingress, nil
}

func (h handler) onChangeNode(key string, node *v1.Node) (*v1.Node, error) {
	if node == nil {
		return node, nil
	}

	if _, ok := node.Labels[nodeSelectorLabel]; !ok {
		return node, nil
	}

	if err := h.updateDaemonSets(); err != nil {
		return node, err
	}

	return node, nil
}

func (h handler) updateDaemonSets() error {
	svc, err := h.services.Cache().Get(h.namespace, constants.GatewayName)
	if err != nil {
		return err
	}

	deepcopy := svc.DeepCopy()
	deepcopy.SystemSpec.PodSpec.NodeSelector = map[string]string{
		nodeSelectorLabel: "true",
	}
	if _, err := h.services.Update(deepcopy); err != nil {
		return err
	}
	return err
}

func (h handler) resolve(namespace, name string, obj runtime.Object) ([]relatedresource.Key, error) {
	switch obj.(type) {
	case *adminv1.ClusterDomain:
		svcs, err := h.services.Cache().List("", labels.NewSelector())
		if err != nil {
			return nil, err
		}
		var keys []relatedresource.Key
		for _, svc := range svcs {
			keys = append(keys, relatedresource.Key{
				Name:      svc.Name,
				Namespace: svc.Namespace,
			})
		}
		return keys, nil
	case *adminv1.PublicDomain:
		return []relatedresource.Key{
			{
				Name:      constants.ClusterDomainName,
				Namespace: h.namespace,
			},
		}, nil
	}

	return nil, nil
}

func (h handler) resolveRouter(namespace, name string, obj runtime.Object) ([]relatedresource.Key, error) {
	apps, err := h.routers.List("", labels.NewSelector())
	if err != nil {
		return nil, err
	}
	var keys []relatedresource.Key
	for _, app := range apps {
		keys = append(keys, relatedresource.Key{
			Name:      app.Name,
			Namespace: app.Namespace,
		})
	}
	return keys, nil
}

func (h handler) resolveApp(namespace, name string, obj runtime.Object) ([]relatedresource.Key, error) {
	switch obj.(type) {
	case *adminv1.ClusterDomain:
		apps, err := h.apps.Cache().List("", labels.NewSelector())
		if err != nil {
			return nil, err
		}
		var keys []relatedresource.Key
		for _, app := range apps {
			keys = append(keys, relatedresource.Key{
				Name:      app.Name,
				Namespace: app.Namespace,
			})
		}
		return keys, nil
	}
	return nil, nil
}

func (h handler) syncDomain(key string, svc *riov1.Service) (*riov1.Service, error) {
	if svc == nil {
		return svc, nil
	}
	if svc.DeletionTimestamp != nil {
		return svc, nil
	}

	clusterDomain, err := h.clusterDomain.Cache().Get(h.namespace, constants.ClusterDomainName)
	if err != nil {
		return svc, err
	}

	updateDomain(svc, clusterDomain)
	return svc, nil
}

func (h handler) syncAppDomain(key string, obj *riov1.App) (*riov1.App, error) {
	if obj == nil {
		return obj, nil
	}
	if obj.DeletionTimestamp != nil {
		return obj, nil
	}

	clusterDomain, err := h.clusterDomain.Cache().Get(h.namespace, constants.ClusterDomainName)
	if err != nil {
		return obj, err
	}

	updateAppDomain(obj, clusterDomain)
	return obj, nil
}

func updateAppDomain(app *riov1.App, clusterDomain *adminv1.ClusterDomain) {
	public := true
	for _, svc := range app.Spec.Revisions {
		if !svc.Public {
			public = false
			break
		}
	}

	protocol := "http"
	if clusterDomain.Status.HTTPSSupported {
		protocol = "https"
	}
	var endpoints []string
	if public && clusterDomain.Status.ClusterDomain != "" {
		endpoints = append(endpoints, fmt.Sprintf("%s://%s", protocol, domains.GetExternalDomain(app.Name, app.Namespace, clusterDomain.Status.ClusterDomain)))
	}
	for _, pd := range app.Status.PublicDomains {
		endpoints = append(endpoints, fmt.Sprintf("%s://%s", protocol, pd))
	}

	app.Status.Endpoints = parse.FormatEndpoint(protocol, endpoints)
}

func updateDomain(service *riov1.Service, clusterDomain *adminv1.ClusterDomain) {
	public := domains.IsPublic(service)

	protocol := "http"
	if clusterDomain.Status.HTTPSSupported {
		protocol = "https"
	}

	var endpoints []string
	if public && clusterDomain.Status.ClusterDomain != "" {
		app, version := services2.AppAndVersion(service)
		endpoints = append(endpoints, fmt.Sprintf("%s://%s", protocol, domains.GetExternalDomain(app+"-"+version, service.Namespace, clusterDomain.Status.ClusterDomain)))
	}

	for _, pd := range service.Status.PublicDomains {
		endpoints = append(endpoints, fmt.Sprintf("%s://%s", protocol, pd))
	}

	service.Status.Endpoints = parse.FormatEndpoint(protocol, endpoints)
}

func (h handler) syncRouterDomain(key string, obj runtime.Object) (runtime.Object, error) {
	if obj == nil {
		return nil, nil
	}

	clusterDomain, err := h.clusterDomain.Cache().Get(h.namespace, constants.ClusterDomainName)
	if err != nil {
		return obj, err
	}

	updateRouterDomain(obj.(*riov1.Router), clusterDomain)

	return obj, nil
}

func updateRouterDomain(router *riov1.Router, clusterDomain *adminv1.ClusterDomain) {
	protocol := "http"
	if clusterDomain.Status.HTTPSSupported {
		protocol = "https"
	}
	router.Status.Endpoints = []string{
		fmt.Sprintf("%s://%s", protocol, domains.GetExternalDomain(router.Name, router.Namespace, clusterDomain.Status.ClusterDomain)),
	}
	for _, pd := range router.Status.PublicDomains {
		router.Status.Endpoints = append(router.Status.Endpoints, fmt.Sprintf("%s://%s", protocol, pd))
	}
	router.Status.Endpoints = parse.FormatEndpoint(protocol, router.Status.Endpoints)
}
