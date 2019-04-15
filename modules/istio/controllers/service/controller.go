package service

import (
	"context"
	"fmt"

	"github.com/rancher/rio/modules/istio/controllers/service/populate"
	"github.com/rancher/rio/modules/istio/pkg/domains"
	v12 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	projectv1controller "github.com/rancher/rio/pkg/generated/controllers/project.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/serviceset"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	serviceDomainUpdate = "service-domain-update"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "routing-service", rContext.Rio.Rio().V1().Service())
	c.Apply = c.Apply.WithStrictCaching().
		WithCacheTypes(rContext.Networking.Networking().V1alpha3().DestinationRule(),
			rContext.Networking.Networking().V1alpha3().VirtualService())

	sh := &serviceHandler{
		systemNamespace:      rContext.Namespace,
		serviceClient:        rContext.Rio.Rio().V1().Service(),
		serviceCache:         rContext.Rio.Rio().V1().Service().Cache(),
		externalServiceCache: rContext.Rio.Rio().V1().ExternalService().Cache(),
		clusterDomainCache:   rContext.Global.Project().V1().ClusterDomain().Cache(),
		publicDomainCache:    rContext.Rio.Rio().V1().PublicDomain().Cache(),
	}

	rContext.Rio.Rio().V1().Service().OnChange(ctx, serviceDomainUpdate, sh.syncDomain)

	c.Populator = sh.populate
	return nil
}

type serviceHandler struct {
	systemNamespace      string
	serviceClient        v1.ServiceClient
	serviceCache         v1.ServiceCache
	externalServiceCache v1.ExternalServiceCache
	clusterDomainCache   projectv1controller.ClusterDomainCache
	publicDomainCache    riov1controller.PublicDomainCache
}

func (s *serviceHandler) populate(obj runtime.Object, namespace *corev1.Namespace, os *objectset.ObjectSet) error {
	service := obj.(*riov1.Service)
	if service.Spec.DisableServiceMesh {
		return nil
	}

	services, err := s.serviceCache.List(service.Namespace, labels.Everything())
	if err != nil {
		return err
	}
	serviceSets, err := serviceset.CollectionServices(services)
	if err != nil {
		return err
	}
	serviceSet, ok := serviceSets[service.Name]
	if !ok {
		return nil
	}

	clusterDomain, err := s.clusterDomainCache.Get(s.systemNamespace, settings.ClusterDomainName)
	if err != nil {
		return err
	}

	if err := populate.DestinationRulesAndVirtualServices(s.systemNamespace, clusterDomain, serviceSet, service, os); err != nil {
		return err
	}

	return err
}

func (s *serviceHandler) syncDomain(key string, svc *riov1.Service) (*riov1.Service, error) {
	if svc == nil {
		return svc, nil
	}

	clusterDomain, err := s.clusterDomainCache.Get(s.systemNamespace, settings.ClusterDomainName)
	if err != nil {
		return svc, err
	}

	updateDomain(svc, clusterDomain)
	return svc, nil
}

func updateDomain(service *riov1.Service, clusterDomain *v12.ClusterDomain) {
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

	if public && clusterDomain.Status.ClusterDomain != "" {
		service.Status.Endpoints = []riov1.Endpoint{
			{
				URL: fmt.Sprintf("%s://%s", protocol, domains.GetExternalDomain(service.Name, service.Namespace, clusterDomain.Status.ClusterDomain)),
			},
		}
	}

	for _, pd := range service.Status.PublicDomains {
		service.Status.Endpoints = append(service.Status.Endpoints, riov1.Endpoint{
			URL: fmt.Sprintf("%s://%s", protocol, pd),
		})
	}
}
