package service

import (
	"context"
	"fmt"

	"github.com/rancher/rio/pkg/name"

	"github.com/rancher/rio/modules/istio/controllers/service/populate"
	"github.com/rancher/rio/modules/istio/pkg/domains"
	"github.com/rancher/rio/modules/system/features/letsencrypt/pkg/issuers"
	projectv1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	corev1controller "github.com/rancher/rio/pkg/generated/controllers/core/v1"
	projectv1controller "github.com/rancher/rio/pkg/generated/controllers/project.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	services2 "github.com/rancher/rio/pkg/services"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	serviceDomainUpdate = "service-domain-update"
	appDomainHandler    = "app-domain-update"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "routing-service", rContext.Rio.Rio().V1().Service())
	c.Apply = c.Apply.WithStrictCaching().
		WithCacheTypes(rContext.Networking.Networking().V1alpha3().DestinationRule(),
			rContext.Networking.Networking().V1alpha3().VirtualService(),
			rContext.Extensions.Extensions().V1beta1().Ingress())

	sh := &serviceHandler{
		systemNamespace:      rContext.Namespace,
		serviceClient:        rContext.Rio.Rio().V1().Service(),
		serviceCache:         rContext.Rio.Rio().V1().Service().Cache(),
		secretCache:          rContext.Core.Core().V1().Secret().Cache(),
		externalServiceCache: rContext.Rio.Rio().V1().ExternalService().Cache(),
		clusterDomainCache:   rContext.Global.Project().V1().ClusterDomain().Cache(),
		publicDomainCache:    rContext.Rio.Rio().V1().PublicDomain().Cache(),
	}

	rContext.Rio.Rio().V1().Service().OnChange(ctx, serviceDomainUpdate, riov1controller.UpdateServiceOnChange(rContext.Rio.Rio().V1().Service().Updater(), sh.syncDomain))
	rContext.Rio.Rio().V1().App().OnChange(ctx, appDomainHandler, riov1controller.UpdateAppOnChange(rContext.Rio.Rio().V1().App().Updater(), sh.syncAppDomain))

	c.Populator = sh.populate
	return nil
}

type serviceHandler struct {
	systemNamespace      string
	serviceClient        v1.ServiceClient
	serviceCache         v1.ServiceCache
	secretCache          corev1controller.SecretCache
	externalServiceCache v1.ExternalServiceCache
	clusterDomainCache   projectv1controller.ClusterDomainCache
	publicDomainCache    riov1controller.PublicDomainCache
}

func (s *serviceHandler) populate(obj runtime.Object, namespace *corev1.Namespace, os *objectset.ObjectSet) error {
	service := obj.(*riov1.Service)
	if service.Spec.DisableServiceMesh {
		return nil
	}

	clusterDomain, err := s.clusterDomainCache.Get(s.systemNamespace, settings.ClusterDomainName)
	if err != nil {
		return err
	}

	if err := populate.DestinationRulesAndVirtualServices(s.systemNamespace, clusterDomain, service, os); err != nil {
		return err
	}

	tls := true
	_, err = s.secretCache.Get(s.systemNamespace, issuers.RioWildcardCerts)
	if err != nil && !errors.IsNotFound(err) {
		tls = false
	} else if errors.IsNotFound(err) {
		return nil
	}

	// generating ingress for revision
	app, version := services2.AppAndVersion(service)
	if err := populate.Ingress(clusterDomain, s.systemNamespace, service.Namespace, name.SafeConcatName(app, version), tls, os); err != nil {
		return err
	}

	return err
}

func (s *serviceHandler) syncDomain(key string, svc *riov1.Service) (*riov1.Service, error) {
	if svc == nil {
		return svc, nil
	}
	if svc.DeletionTimestamp != nil {
		return svc, nil
	}

	clusterDomain, err := s.clusterDomainCache.Get(s.systemNamespace, settings.ClusterDomainName)
	if err != nil {
		return svc, err
	}

	updateDomain(svc, clusterDomain)
	return svc, nil
}

func (s *serviceHandler) syncAppDomain(key string, obj *riov1.App) (*riov1.App, error) {
	if obj == nil {
		return obj, nil
	}
	if obj.DeletionTimestamp != nil {
		return obj, nil
	}

	clusterDomain, err := s.clusterDomainCache.Get(s.systemNamespace, settings.ClusterDomainName)
	if err != nil {
		return obj, err
	}

	updateAppDomain(obj, clusterDomain)
	return obj, nil
}

func updateAppDomain(app *riov1.App, clusterDomain *projectv1.ClusterDomain) {
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
	app.Status.Endpoints = endpoints
}

func updateDomain(service *riov1.Service, clusterDomain *projectv1.ClusterDomain) {
	public := false
	for _, port := range service.Spec.Ports {
		if !port.InternalOnly {
			public = true
			break
		}
	}

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
	service.Status.Endpoints = endpoints
}
