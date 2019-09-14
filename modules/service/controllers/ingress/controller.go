package ingress

import (
	"context"

	"github.com/rancher/rio/modules/service/controllers/ingress/populate"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	adminv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/objectset"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	serviceController := stackobject.NewGeneratingController(ctx, rContext, "ingress-service", rContext.Rio.Rio().V1().Service())
	serviceController.Apply = serviceController.Apply.WithStrictCaching().
		WithCacheTypes(rContext.K8sNetworking.Networking().V1beta1().Ingress())

	appController := stackobject.NewGeneratingController(ctx, rContext, "ingress-app", rContext.Rio.Rio().V1().App())
	appController.Apply = appController.Apply.WithStrictCaching().
		WithCacheTypes(rContext.K8sNetworking.Networking().V1beta1().Ingress())

	routerController := stackobject.NewGeneratingController(ctx, rContext, "ingress-routeset", rContext.Rio.Rio().V1().Router())
	routerController.Apply = routerController.Apply.WithStrictCaching().
		WithCacheTypes(rContext.K8sNetworking.Networking().V1beta1().Ingress())

	publicdomainController := stackobject.NewGeneratingController(ctx, rContext, "ingress-publicdomain", rContext.Global.Admin().V1().PublicDomain())
	publicdomainController.Apply = publicdomainController.Apply.WithCacheTypes(rContext.Networking.Networking().V1alpha3().DestinationRule(),
		rContext.K8sNetworking.Networking().V1beta1().Ingress())

	sh := &handler{
		systemNamespace:      rContext.Namespace,
		serviceClient:        rContext.Rio.Rio().V1().Service(),
		serviceCache:         rContext.Rio.Rio().V1().Service().Cache(),
		secretCache:          rContext.Core.Core().V1().Secret().Cache(),
		externalServiceCache: rContext.Rio.Rio().V1().ExternalService().Cache(),
		clusterDomainCache:   rContext.Global.Admin().V1().ClusterDomain().Cache(),
		publicDomainCache:    rContext.Global.Admin().V1().PublicDomain().Cache(),
		routesetCache:        rContext.Rio.Rio().V1().Router().Cache(),
	}

	serviceController.Populator = sh.populateService
	appController.Populator = sh.populateApp
	routerController.Populator = sh.populateRouter
	publicdomainController.Populator = sh.populatePublicDomain

	return nil
}

type handler struct {
	systemNamespace      string
	serviceClient        v1.ServiceClient
	serviceCache         v1.ServiceCache
	secretCache          corev1controller.SecretCache
	externalServiceCache v1.ExternalServiceCache
	clusterDomainCache   adminv1controller.ClusterDomainCache
	publicDomainCache    adminv1controller.PublicDomainCache
	routesetCache        riov1controller.RouterCache
}

func (h handler) populateService(obj runtime.Object, namespace *corev1.Namespace, os *objectset.ObjectSet) error {
	service := obj.(*riov1.Service)
	if service.Spec.DisableServiceMesh {
		return nil
	}

	clusterDomain, err := h.clusterDomainCache.Get(h.systemNamespace, constants.ClusterDomainName)
	if err != nil {
		return err
	}

	if clusterDomain.Status.ClusterDomain != "" && constants.InstallMode == constants.InstallModeIngress {
		populate.Ingress(h.systemNamespace, clusterDomain.Status.ClusterDomain, clusterDomain.Spec.SecretRef.Name, false, service, os)
	}

	return err
}

func (h handler) populateApp(obj runtime.Object, namespace *corev1.Namespace, os *objectset.ObjectSet) error {
	app := obj.(*riov1.App)
	if app == nil {
		return nil
	}

	clusterDomain, err := h.clusterDomainCache.Get(h.systemNamespace, constants.ClusterDomainName)
	if err != nil {
		return err
	}

	if len(app.Spec.Revisions) == 0 {
		return nil
	}

	public := false
	for _, rev := range app.Spec.Revisions {
		if rev.Public {
			public = true
		}
	}
	if !public {
		return nil
	}

	var revision *riov1.Service
	for i := len(app.Spec.Revisions) - 1; i >= 0; i-- {
		revision, err = h.serviceCache.Get(app.Namespace, app.Spec.Revisions[i].ServiceName)
		if err != nil && !errors.IsNotFound(err) {
			return err
		} else if errors.IsNotFound(err) {
			continue
		}
		break
	}
	if revision == nil {
		return nil
	}

	if clusterDomain.Status.ClusterDomain != "" && constants.InstallMode == constants.InstallModeIngress {
		populate.Ingress(h.systemNamespace, clusterDomain.Status.ClusterDomain, clusterDomain.Spec.SecretRef.Name, true, revision, os)
	}

	return nil
}

func (h handler) populateRouter(obj runtime.Object, ns *corev1.Namespace, os *objectset.ObjectSet) error {
	routeSet := obj.(*riov1.Router)

	clusterDomain, err := h.clusterDomainCache.Get(h.systemNamespace, constants.ClusterDomainName)
	if err != nil {
		return err
	}
	domain := clusterDomain.Status.ClusterDomain
	if domain == "" {
		return nil
	}

	if constants.InstallMode == constants.InstallModeIngress {
		populate.IngressForRouter(h.systemNamespace, domain, clusterDomain.Spec.SecretRef.Name, routeSet, os)
	}

	return nil
}

func (h handler) populatePublicDomain(obj runtime.Object, ns *corev1.Namespace, os *objectset.ObjectSet) error {
	pd := obj.(*adminv1.PublicDomain)
	if constants.InstallMode == constants.InstallModeIngress {
		populate.IngressForPublicDomain(h.systemNamespace, pd, os)
	}
	return nil
}
