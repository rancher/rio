package istio

import (
	"context"
	"fmt"

	"github.com/knative/pkg/apis/istio/v1alpha3"
	errors2 "github.com/pkg/errors"
	"github.com/rancher/rio/exclude/pkg/settings"
	"github.com/rancher/rio/modules/istio/controllers/istio/populate"
	"github.com/rancher/rio/modules/istio/pkg/istio/config"
	projectv1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	v12 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	corev1controller "github.com/rancher/rio/pkg/generated/controllers/core/v1"
	projectv1controller "github.com/rancher/rio/pkg/generated/controllers/project.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/apply/injectors"
	"github.com/rancher/wrangler/pkg/objectset"
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
	istioDeploy   = "istio-deploy"
	istioStack    = "istio-stack"
)

var (
	addressTypes = []v1.NodeAddressType{
		v1.NodeExternalIP,
		v1.NodeInternalIP,
	}

	evalTrigger trigger.Trigger

	meshTemplate = `disablePolicyChecks: true
ingressControllerMode: "OFF"
authPolicy: NONE
rdsRefreshDelay: 10s
mixerReportServer: %s.rio-cloud.svc.cluster.local:9091
outboundTrafficPolicy:
  mode: ALLOW_ANY
defaultConfig:
  discoveryRefreshDelay: 10s
  connectTimeout: 30s
  configPath: "/etc/istio/proxy"
  binaryPath: "/usr/local/bin/envoy"
  serviceCluster: istio-proxy
  drainDuration: 45s
  parentShutdownDuration: 1m0s
  interceptionMode: REDIRECT
  proxyAdminPort: 15000
  controlPlaneAuthPolicy: NONE
  discoveryAddress: %s:15007`
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
		gatewayApply: rContext.Apply.WithSetID(istioStack).
			WithCacheTypes(rContext.Networking.Networking().V1alpha3().Gateway()),
		serviceApply:      rContext.Apply.WithSetID(istioInjector).WithInjectorName(istioInjector),
		publicDomainCache: rContext.Global.Project().V1().PublicDomain().Cache(),
		clusterDomain:     rContext.Global.Project().V1().ClusterDomain(),
		secretCache:       rContext.Core.Core().V1().Secret().Cache(),
		nodeCache:         rContext.Core.Core().V1().Node().Cache(),
	}

	evalTrigger = trigger.New(rContext.Networking.Networking().V1alpha3().VirtualService())
	evalTrigger.OnTrigger(ctx, istioDeploy, s.sync)

	relatedresource.Watch(ctx, istioDeploy,
		resolve,
		rContext.Networking.Networking().V1alpha3().VirtualService(),
		rContext.Networking.Networking().V1alpha3().VirtualService(),
		rContext.Global.Project().V1().ClusterDomain())

	rContext.Core.Core().V1().Endpoints().OnChange(ctx, "istio-endpoints", s.syncEndpoint)

	return nil
}

func resolve(namespace, name string, obj runtime.Object) ([]relatedresource.Key, error) {
	switch obj.(type) {
	case *projectv1.ClusterDomain:
		return []relatedresource.Key{evalTrigger.Key()}, nil
	case *v1alpha3.VirtualService:
		return []relatedresource.Key{evalTrigger.Key()}, nil
	}

	return nil, nil
}

type istioDeployController struct {
	namespace         string
	gatewayApply      apply.Apply
	serviceApply      apply.Apply
	publicDomainCache projectv1controller.PublicDomainCache
	clusterDomain     projectv1controller.ClusterDomainController
	secretCache       corev1controller.SecretCache
	nodeCache         corev1controller.NodeCache
}

/*
	sync creates istio components and ingress gateway when changes happen
*/
func (i *istioDeployController) sync() error {
	output := objectset.NewObjectSet()
	if err := populate.PopulateIstioDeploy(i.namespace, output); err != nil {
		output.AddErr(err)
	}
	if err := i.serviceApply.Apply(output); err != nil {
		return err
	}

	pds, err := i.publicDomainCache.List("", labels.Everything())
	if err != nil {
		return err
	}

	clusterDomain, err := i.clusterDomain.Cache().Get(i.namespace, settings.ClusterDomainName)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	secret, err := i.secretCache.Get(i.namespace, settings.GatewaySecretName)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	publicdomainSecrets := map[string]*v1.Secret{}
	for _, pd := range pds {
		key := fmt.Sprintf("%s/%s", pd.Namespace, pd.Name)
		secret, err := i.secretCache.Get(i.namespace, key)
		if err == nil {
			publicdomainSecrets[key] = secret
		}
	}

	os := populate.Istio(i.namespace, clusterDomain, pds, publicdomainSecrets, secret)
	return i.gatewayApply.Apply(os)
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

func (i *istioDeployController) injectService(key string, svc *riov1.Service) (*riov1.Service, error) {
	if svc.Spec.DisableServiceMesh {
		return svc, nil
	}

	output := objectset.NewObjectSet()
	output.Add(svc)
	err := i.serviceApply.Apply(output)
	return svc, err
}

func (i *istioDeployController) syncEndpoint(key string, endpoint *corev1.Endpoints) (*corev1.Endpoints, error) {
	if endpoint.Namespace != i.namespace || endpoint.Name != settings.IstioGatewayDeploy {
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

	clusterDomain, err := i.clusterDomain.Cache().Get(i.namespace, settings.ClusterDomainName)
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
	cf := config.NewConfigFactory(ctx, rContext.Core.Core().V1().ConfigMap(),
		settings.IstioStackName,
		settings.IstionConfigMapName,
		settings.IstionConfigMapKey)
	injector := config.NewIstioInjector(cf)
	injectors.Register(istioInjector, injector.Inject)

	meshConfig := constructors.NewConfigMap(rContext.Namespace, settings.IstionConfigMapName, v1.ConfigMap{
		Data: map[string]string{
			settings.IstionConfigMapKey: fmt.Sprintf(meshTemplate, rContext.Namespace, settings.IstioPilot),
		},
	})

	if _, err := rContext.Core.Core().V1().ConfigMap().Create(meshConfig); err != nil && !errors.IsAlreadyExists(err) {
		return errors2.Wrap(err, "failed to create istio mesh config")
	}

	return nil
}

func ensureClusterDomain(ns string, clusterDomain projectv1controller.ClusterDomainClient) error {
	_, err := clusterDomain.Get(ns, settings.ClusterDomainName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err := clusterDomain.Create(v12.NewClusterDomain(ns, settings.ClusterDomainName, v12.ClusterDomain{}))
		return err
	}
	return err
}
