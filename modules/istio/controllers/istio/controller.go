package istio

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/rancher/rio/modules/istio/controllers/istio/populate"
	"github.com/rancher/rio/modules/istio/pkg/istio/config"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	adminv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	appsv1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/apps/v1"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/apply/injectors"
	"github.com/rancher/wrangler/pkg/objectset"
	"github.com/rancher/wrangler/pkg/relatedresource"
	"github.com/rancher/wrangler/pkg/trigger"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	istioInjector = "istio-injecter"
	istioStack    = "istio-stack"

	nodeSelectorLabel = "rio.cattle.io/gateway"
	indexName         = "nodeEndpointIndexer"
)

var (
	addressTypes = []v1.NodeAddressType{
		v1.NodeExternalIP,
		v1.NodeInternalIP,
	}

	evalTrigger trigger.Trigger
)

func Register(ctx context.Context, rContext *types.Context) error {
	if err := ensureClusterDomain(rContext.Namespace, rContext.Global.Admin().V1().ClusterDomain()); err != nil {
		return err
	}

	if err := setupConfigmapAndInjectors(ctx, rContext); err != nil {
		return err
	}

	if err := enqueueServicesForInject(rContext.Rio.Rio().V1().Service()); err != nil {
		return err
	}

	s := &istioDeployController{
		namespace: rContext.Namespace,
		apply:     rContext.Apply,
		gatewayApply: rContext.Apply.WithSetID(istioStack).
			WithCacheTypes(rContext.Networking.Networking().V1alpha3().Gateway()),
		serviceApply:      rContext.Apply.WithSetID(istioInjector).WithInjectorName(istioInjector),
		apps:              rContext.Rio.Rio().V1().App(),
		services:          rContext.Rio.Rio().V1().Service(),
		publicDomainCache: rContext.Global.Admin().V1().PublicDomain().Cache(),
		clusterDomain:     rContext.Global.Admin().V1().ClusterDomain(),
		secretCache:       rContext.Core.Core().V1().Secret().Cache(),
		nodeCache:         rContext.Core.Core().V1().Node().Cache(),
		endpointCache:     rContext.Core.Core().V1().Endpoints().Cache(),
		daemonsets:        rContext.Apps.Apps().V1().DaemonSet(),
	}

	relatedresource.Watch(ctx, "app-clusterdomain", s.resolveApp,
		rContext.Rio.Rio().V1().App(),
		rContext.Global.Admin().V1().ClusterDomain())

	relatedresource.Watch(ctx, "publicdomain-clusterdomain", s.resolve,
		rContext.Rio.Rio().V1().Service(),
		rContext.Global.Admin().V1().ClusterDomain())

	relatedresource.Watch(ctx, "cluster-domain-service", s.resolve,
		rContext.Global.Admin().V1().ClusterDomain(),
		rContext.Global.Admin().V1().PublicDomain())

	relatedresource.Watch(ctx, "node-enpoint", s.resolveEndpoint,
		rContext.Core.Core().V1().Endpoints(),
		rContext.Core.Core().V1().Node())

	rContext.Core.Core().V1().Endpoints().Cache().AddIndexer(indexName, s.indexEPByNode)

	switch {
	case constants.UseIPAddress != "":
		addresses := strings.Split(constants.UseIPAddress, ",")
		if err := s.updateClusterDomain(addresses); err != nil {
			return err
		}
	case !constants.UseHostPort:
		rContext.Core.Core().V1().Service().OnChange(ctx, "istio-endpoints-serviceloadbalancer", s.syncServiceLoadbalancer)
	default:
		rContext.Core.Core().V1().Endpoints().OnChange(ctx, "istio-endpoints", s.syncEndpoint)
	}

	rContext.Core.Core().V1().Service().OnChange(ctx, "rdns-subdomain", s.syncSubdomain)
	rContext.Global.Admin().V1().ClusterDomain().OnChange(ctx, "clusterdomain-gateway", s.syncGateway)

	rContext.Core.Core().V1().Node().OnChange(ctx, "gateway-daemonset-update", s.onChangeNode)

	return nil
}

type istioDeployController struct {
	namespace         string
	apply             apply.Apply
	gatewayApply      apply.Apply
	serviceApply      apply.Apply
	apps              riov1controller.AppController
	services          riov1controller.ServiceController
	publicDomainCache adminv1controller.PublicDomainCache
	clusterDomain     adminv1controller.ClusterDomainController
	secretCache       corev1controller.SecretCache
	nodeCache         corev1controller.NodeCache
	endpointCache     corev1controller.EndpointsCache
	daemonsets        appsv1controller.DaemonSetController
}

func (i *istioDeployController) onChangeNode(key string, node *corev1.Node) (*corev1.Node, error) {
	if _, ok := node.Labels[nodeSelectorLabel]; !ok {
		return node, nil
	}

	if err := i.updateDaemonSets(); err != nil {
		return node, err
	}

	return node, nil
}

func (i *istioDeployController) updateDaemonSets() error {
	svc, err := i.services.Cache().Get(i.namespace, constants.IstioGateway)
	if err != nil {
		return err
	}

	deepcopy := svc.DeepCopy()
	deepcopy.SystemSpec.PodSpec.NodeSelector = map[string]string{
		nodeSelectorLabel: "true",
	}
	if _, err := i.services.Update(deepcopy); err != nil {
		return err
	}
	return err
}

func (i istioDeployController) syncServiceLoadbalancer(key string, obj *v1.Service) (*v1.Service, error) {
	if obj == nil {
		return obj, nil
	}

	if obj.Spec.Selector["app"] != constants.IstioGateway || obj.Namespace != i.namespace || obj.Spec.Type != v1.ServiceTypeLoadBalancer {
		return obj, nil
	}

	var address []string
	for _, ingress := range obj.Status.LoadBalancer.Ingress {
		if ingress.Hostname == "localhost" {
			ingress.IP = "127.0.0.1"
		}
		address = append(address, ingress.IP)
	}

	if err := i.updateClusterDomain(address); err != nil {
		return obj, err
	}
	return obj, nil
}

func (i istioDeployController) syncGateway(key string, obj *adminv1.ClusterDomain) (*adminv1.ClusterDomain, error) {
	if obj == nil || obj.DeletionTimestamp != nil || obj.Name != constants.ClusterDomainName {
		return obj, nil
	}

	os := objectset.NewObjectSet()
	domain := ""
	if obj.Status.ClusterDomain != "" {
		domain = fmt.Sprintf("*.%s", obj.Status.ClusterDomain)
	}

	publicdomains, err := i.publicDomainCache.List("", labels.NewSelector())
	if err != nil {
		return obj, err
	}
	populate.Gateway(i.namespace, domain, publicdomains, os)
	return obj, i.apply.WithSetID("istio-gateway").Apply(os)
}

func enqueueServicesForInject(controller riov1controller.ServiceController) error {
	svcs, err := controller.Cache().List("", labels.Everything())
	if err != nil {
		return err
	}

	for _, svc := range svcs {
		controller.Enqueue(svc.Namespace, svc.Name)
	}
	return nil
}

func (i *istioDeployController) resolve(namespace, name string, obj runtime.Object) ([]relatedresource.Key, error) {
	switch obj.(type) {
	case *adminv1.ClusterDomain:
		svcs, err := i.services.Cache().List("", labels.NewSelector())
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
				Namespace: i.namespace,
			},
		}, nil
	}

	return nil, nil
}

func (i *istioDeployController) resolveApp(namespace, name string, obj runtime.Object) ([]relatedresource.Key, error) {
	switch obj.(type) {
	case *adminv1.ClusterDomain:
		apps, err := i.apps.Cache().List("", labels.NewSelector())
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

func (i *istioDeployController) resolveEndpoint(namespace, name string, obj runtime.Object) ([]relatedresource.Key, error) {
	switch obj.(type) {
	case *corev1.Node:
		eps, err := i.endpointCache.GetByIndex(indexName, obj.(*corev1.Node).Name)
		if err != nil || len(eps) == 0 {
			return nil, err
		}
		return []relatedresource.Key{
			{
				Namespace: eps[0].Namespace,
				Name:      eps[0].Name,
			},
		}, nil

	}
	return nil, nil
}

func (i *istioDeployController) syncEndpoint(key string, endpoint *corev1.Endpoints) (*corev1.Endpoints, error) {
	if endpoint == nil {
		return nil, nil
	}
	if endpoint.Namespace != i.namespace || endpoint.Name != constants.IstioGateway {
		return endpoint, nil
	}

	var ips []string
	for _, subset := range endpoint.Subsets {
		for _, addr := range subset.Addresses {
			if addr.NodeName == nil {
				continue
			}

			node, err := i.nodeCache.Get(*addr.NodeName)
			if err != nil {
				return nil, err
			}

			nodeIP := getNodeIP(node)
			if nodeIP != "" {
				ips = append(ips, nodeIP)
			}
		}
	}
	if err := i.updateClusterDomain(ips); err != nil {
		return endpoint, err
	}
	return endpoint, nil
}

func (i istioDeployController) updateClusterDomain(addresses []string) error {
	clusterDomain, err := i.clusterDomain.Cache().Get(i.namespace, constants.ClusterDomainName)
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

	if _, err := i.clusterDomain.Update(deepcopy); err != nil {
		return err
	}

	return err
}

func (i *istioDeployController) indexEPByNode(ep *corev1.Endpoints) ([]string, error) {
	if ep.Namespace != i.namespace || ep.Name != constants.IstioGateway {
		return nil, nil
	}

	var result []string

	for _, subset := range ep.Subsets {
		for _, addr := range subset.Addresses {
			if addr.NodeName != nil {
				result = append(result, *addr.NodeName)
			}
		}
	}

	return result, nil
}

func (i *istioDeployController) syncSubdomain(key string, service *corev1.Service) (*corev1.Service, error) {
	if service == nil {
		return service, nil
	}
	clusterDomain, err := i.clusterDomain.Cache().Get(i.namespace, constants.ClusterDomainName)
	if err != nil {
		return service, err
	}

	if service.Labels["request-subdomain"] == "true" && service.Spec.ClusterIP != "" {
		deepcopy := clusterDomain.DeepCopy()
		found := false
		appName := service.Spec.Selector["app"]
		subdomainName := fmt.Sprintf("%s-%s", appName, service.Namespace)
		for i, sd := range deepcopy.Spec.Subdomains {
			if sd.Name == subdomainName {
				found = true
				deepcopy.Spec.Subdomains[i].Addresses = []adminv1.Address{
					{
						IP: service.Spec.ClusterIP,
					},
				}
			}
		}

		if !found {
			deepcopy.Spec.Subdomains = append(deepcopy.Spec.Subdomains, adminv1.Subdomain{
				Name: subdomainName,
				Addresses: []adminv1.Address{
					{
						IP: service.Spec.ClusterIP,
					},
				},
			})
		}

		if _, err := i.clusterDomain.Update(deepcopy); err != nil {
			return service, err
		}
	}

	return service, nil
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

func setupConfigmapAndInjectors(ctx context.Context, rContext *types.Context) error {
	cm, err := rContext.Core.Core().V1().ConfigMap().Get(rContext.Namespace, constants.IstionConfigMapName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	meshConfig, template, err := config.DoConfigAndTemplate(cm.Data[constants.IstioMeshConfigKey], cm.Data[constants.IstioSidecarTemplateName])
	if err != nil {
		return err
	}

	injector := config.NewIstioInjector(meshConfig, template)
	injectors.Register(istioInjector, injector.Inject)
	return nil
}

func ensureClusterDomain(ns string, clusterDomain adminv1controller.ClusterDomainClient) error {
	_, err := clusterDomain.Get(ns, constants.ClusterDomainName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err := clusterDomain.Create(adminv1.NewClusterDomain(ns, constants.ClusterDomainName, adminv1.ClusterDomain{
			Spec: adminv1.ClusterDomainSpec{
				SecretRef: v1.SecretReference{
					Namespace: ns,
					Name:      constants.GatewaySecretName,
				},
			},
		}))
		return err
	}
	return err
}
