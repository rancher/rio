package istio

import (
	"context"
	"fmt"

	"github.com/rancher/rio/modules/istio/controllers/istio/populate"
	"github.com/rancher/rio/modules/istio/pkg/istio/config"
	projectv1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	corev1controller "github.com/rancher/rio/pkg/generated/controllers/core/v1"
	projectv1controller "github.com/rancher/rio/pkg/generated/controllers/project.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/apply/injectors"
	"github.com/rancher/wrangler/pkg/relatedresource"
	"github.com/rancher/wrangler/pkg/trigger"
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

	indexName = "nodeEndpointIndexer"
)

var (
	addressTypes = []v1.NodeAddressType{
		v1.NodeExternalIP,
		v1.NodeInternalIP,
	}

	evalTrigger trigger.Trigger
)

func Register(ctx context.Context, rContext *types.Context) error {
	if err := ensureClusterDomain(rContext.Namespace, rContext.Global.Project().V1().ClusterDomain()); err != nil {
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
		services:          rContext.Rio.Rio().V1().Service(),
		publicDomainCache: rContext.Rio.Rio().V1().PublicDomain().Cache(),
		clusterDomain:     rContext.Global.Project().V1().ClusterDomain(),
		secretCache:       rContext.Core.Core().V1().Secret().Cache(),
		nodeCache:         rContext.Core.Core().V1().Node().Cache(),
		endpointCache:     rContext.Core.Core().V1().Endpoints().Cache(),
	}

	if err := s.gatewayApply.Apply(populate.Istio(s.namespace)); err != nil {
		return err
	}

	relatedresource.Watch(ctx, "cluster-domain-service", s.resolve,
		rContext.Rio.Rio().V1().Service(),
		rContext.Global.Project().V1().ClusterDomain(),
	)
	relatedresource.Watch(ctx, "node-enpoint", s.resolveEndpoint,
		rContext.Core.Core().V1().Endpoints(),
		rContext.Core.Core().V1().Node())

	rContext.Core.Core().V1().Endpoints().Cache().AddIndexer(indexName, s.indexEPByNode)
	rContext.Core.Core().V1().Endpoints().OnChange(ctx, "istio-endpoints", s.syncEndpoint)

	rContext.Core.Core().V1().Service().OnChange(ctx, "rdns-subdomain", s.syncSubdomain)

	return nil
}

type istioDeployController struct {
	namespace         string
	apply             apply.Apply
	gatewayApply      apply.Apply
	serviceApply      apply.Apply
	services          riov1controller.ServiceController
	publicDomainCache riov1controller.PublicDomainCache
	clusterDomain     projectv1controller.ClusterDomainController
	secretCache       corev1controller.SecretCache
	nodeCache         corev1controller.NodeCache
	endpointCache     corev1controller.EndpointsCache
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
	case *projectv1.ClusterDomain:
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
	if endpoint.Namespace != i.namespace || endpoint.Name != constants.IstioGatewayDeploy {
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

	clusterDomain, err := i.clusterDomain.Cache().Get(i.namespace, constants.ClusterDomainName)
	if err != nil && !errors.IsNotFound(err) {
		return endpoint, err
	}

	if clusterDomain == nil {
		return endpoint, nil
	}

	deepcopy := clusterDomain.DeepCopy()
	var address []projectv1.Address
	for _, ip := range ips {
		address = append(address, projectv1.Address{IP: ip})
	}
	deepcopy.Spec.Addresses = address

	if _, err := i.clusterDomain.Update(deepcopy); err != nil {
		return endpoint, err
	}

	return endpoint, nil
}

func (i *istioDeployController) indexEPByNode(ep *corev1.Endpoints) ([]string, error) {
	if ep.Namespace != i.namespace || ep.Name != constants.IstioGatewayDeploy {
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
				deepcopy.Spec.Subdomains[i].Addresses = []projectv1.Address{
					{
						IP: service.Spec.ClusterIP,
					},
				}
			}
		}

		if !found {
			deepcopy.Spec.Subdomains = append(deepcopy.Spec.Subdomains, projectv1.Subdomain{
				Name: subdomainName,
				Addresses: []projectv1.Address{
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

func ensureClusterDomain(ns string, clusterDomain projectv1controller.ClusterDomainClient) error {
	_, err := clusterDomain.Get(ns, constants.ClusterDomainName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err := clusterDomain.Create(projectv1.NewClusterDomain(ns, constants.ClusterDomainName, projectv1.ClusterDomain{
			Spec: projectv1.ClusterDomainSpec{
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
