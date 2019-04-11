package service

import (
	"context"
	"fmt"

	v12 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/rancher/rio/modules/istio/pkg/domains"

	"github.com/rancher/rio/exclude/pkg/settings"

	"github.com/rancher/rio/modules/istio/controllers/service/populate"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	projectv1controller "github.com/rancher/rio/pkg/generated/controllers/project.rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
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
		clusterDomain:        rContext.Global.Project().V1().ClusterDomain(),
	}

	// as a side effect of the stack service controller, all changes to revisions will enqueue the parent service

	c.Populator = sh.populate
	return nil
}

type serviceHandler struct {
	systemNamespace      string
	serviceClient        v1.ServiceClient
	serviceCache         v1.ServiceCache
	externalServiceCache v1.ExternalServiceCache
	clusterDomain        projectv1controller.ClusterDomainController
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

	clusterDomain, err := s.ensureClusterDomain()
	if err != nil {
		return err
	}

	if err := populate.DestinationRulesAndVirtualServices(s.systemNamespace, clusterDomain, services, service, os); err != nil {
		return err
	}

	public := false
	for _, port := range service.Spec.Ports {
		if port.Publish == true {
			public = true
			break
		}
	}

	if public && clusterDomain.Status.ClusterDomain != "" {
		protocol := "http"
		if clusterDomain.Status.HTTPSSupported {
			protocol = "https"
		}
		deepcopy := service.DeepCopy()
		deepcopy.Status.Endpoints = append(deepcopy.Status.Endpoints, riov1.Endpoint{
			URL: fmt.Sprintf("%s://%s", protocol, domains.GetExternalDomain(service.Name, service.Namespace, clusterDomain.Status.ClusterDomain)),
		})
		if _, err := s.serviceClient.Update(deepcopy); err != nil {
			return err
		}
	} else {
		deepcopy := service.DeepCopy()
		deepcopy.Status.Endpoints = nil
		if _, err := s.serviceClient.Update(deepcopy); err != nil {
			return err
		}
	}

	return nil
}

func (s *serviceHandler) ensureClusterDomain() (*v12.ClusterDomain, error) {
	clusterDomain, err := s.clusterDomain.Cache().Get(s.systemNamespace, settings.ClusterDomainName)
	if errors.IsNotFound(err) {
		return s.clusterDomain.Create(v12.NewClusterDomain(s.systemNamespace, settings.ClusterDomainName, v12.ClusterDomain{}))
	}
	return clusterDomain, err
}
