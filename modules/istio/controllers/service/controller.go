package service

import (
	"context"
	"fmt"

	"github.com/rancher/rio/exclude/pkg/settings"
	"github.com/rancher/rio/modules/istio/controllers/service/populate"
	"github.com/rancher/rio/modules/istio/pkg/domains"
	v12 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	projectv1controller "github.com/rancher/rio/pkg/generated/controllers/project.rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	"github.com/rancher/wrangler/pkg/relatedresource"
	"github.com/rancher/wrangler/pkg/trigger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	domainTrigger trigger.Trigger
)

const (
	domainUpdate = "domain-update"
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
		publicDomainCache:    rContext.Global.Project().V1().PublicDomain().Cache(),
	}

	domainTrigger = trigger.New(rContext.Rio.Rio().V1().Service())
	domainTrigger.OnTrigger(ctx, domainUpdate, sh.syncDomain)

	relatedresource.Watch(ctx, domainUpdate,
		resolve,
		rContext.Rio.Rio().V1().Service(),
		rContext.Global.Project().V1().ClusterDomain())

	// as a side effect of the stack service controller, all changes to revisions will enqueue the parent service

	c.Populator = sh.populate
	return nil
}

func resolve(namespace, name string, obj runtime.Object) ([]relatedresource.Key, error) {
	switch obj.(type) {
	case *v12.ClusterDomain:
		return []relatedresource.Key{domainTrigger.Key()}, nil
	}
	return nil, nil
}

type serviceHandler struct {
	systemNamespace      string
	serviceClient        v1.ServiceClient
	serviceCache         v1.ServiceCache
	externalServiceCache v1.ExternalServiceCache
	clusterDomainCache   projectv1controller.ClusterDomainCache
	publicDomainCache    projectv1controller.PublicDomainCache
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

	clusterDomain, err := s.clusterDomainCache.Get(s.systemNamespace, settings.ClusterDomainName)
	if err != nil {
		return err
	}

	publicDomains, err := s.publicDomainCache.List(s.systemNamespace, labels.NewSelector())
	if err != nil {
		return err
	}

	var publicDomainsForService []*v12.PublicDomain
	for _, publicDomain := range publicDomains {
		if publicDomain.Spec.TargetServiceName == service.Name {
			publicDomainsForService = append(publicDomainsForService, publicDomain)
		}
	}

	if err := populate.DestinationRulesAndVirtualServices(s.systemNamespace, clusterDomain, publicDomainsForService, services, service, os); err != nil {
		return err
	}

	deepcopy := service.DeepCopy()
	updateDomain(deepcopy, clusterDomain)
	_, err = s.serviceClient.Update(deepcopy)
	return err
}

func (s *serviceHandler) syncDomain() error {
	clusterDomain, err := s.clusterDomainCache.Get(s.systemNamespace, settings.ClusterDomainName)
	if err != nil {
		return err
	}

	services, err := s.serviceCache.List(s.systemNamespace, labels.Everything())
	if err != nil {
		return err
	}

	for _, service := range services {
		deepcopy := service.DeepCopy()
		updateDomain(deepcopy, clusterDomain)
		_, err := s.serviceClient.Update(deepcopy)
		return err
	}

	return nil
}

func updateDomain(service *riov1.Service, clusterDomain *v12.ClusterDomain) {
	public := false
	for _, port := range service.Spec.Ports {
		if !port.InternalOnly {
			public = true
			break
		}
	}

	if public && clusterDomain.Status.ClusterDomain != "" {
		protocol := "http"
		if clusterDomain.Status.HTTPSSupported {
			protocol = "https"
		}
		service.Status.Endpoints = []riov1.Endpoint{
			{
				URL: fmt.Sprintf("%s://%s", protocol, domains.GetExternalDomain(service.Name, service.Namespace, clusterDomain.Status.ClusterDomain)),
			},
		}
	} else {
		service.Status.Endpoints = nil
	}
}
