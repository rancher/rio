package endpoints

import (
	"context"
	"fmt"
	"sort"

	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"

	"github.com/rancher/rio/pkg/indexes"

	"github.com/rancher/rio/pkg/services"
	name2 "github.com/rancher/wrangler/pkg/name"

	"github.com/rancher/rio/modules/service/pkg/domains"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	adminv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/relatedresource"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

type Resolver struct {
	namespace          string
	clusterDomainCache adminv1controller.ClusterDomainCache
	publicDomainCache  adminv1controller.PublicDomainCache
}

func NewResolver(ctx context.Context, namespace string,
	enqueuer relatedresource.Enqueuer,
	serviceCache riov1controller.ServiceCache,
	clusterDomainController adminv1controller.ClusterDomainController,
	publicDomainController adminv1controller.PublicDomainController) *Resolver {

	relatedresource.Watch(ctx,
		"endpoint-resolver",
		func(namespace, name string, obj runtime.Object) (keys []relatedresource.Key, err error) {
			return lookupServices(namespace, name, serviceCache, obj)
		},
		enqueuer,
		clusterDomainController,
		publicDomainController)

	return &Resolver{
		namespace:          namespace,
		clusterDomainCache: clusterDomainController.Cache(),
		publicDomainCache:  publicDomainController.Cache(),
	}
}

func lookupServices(namespace, name string, serviceCache riov1controller.ServiceCache, obj runtime.Object) ([]relatedresource.Key, error) {
	switch obj := obj.(type) {
	case *adminv1.PublicDomain:
		if obj.Spec.TargetApp == "" && obj.Spec.TargetRouter == "" {
			return nil, nil
		}

		ns := obj.Spec.TargetNamespace
		if ns == "" {
			ns = obj.Namespace
		}

		var keys []relatedresource.Key

		target := obj.Spec.TargetApp
		if target == "" {
			target = obj.Spec.TargetRouter
		}

		// router will always be app == name
		keys = append(keys, relatedresource.Key{
			Namespace: ns,
			Name:      target,
		})

		// keys for services
		apps, err := serviceCache.GetByIndex(indexes.ServiceByApp, fmt.Sprintf("%s/%s", ns, target))
		if err != nil {
			return nil, err
		}

		for _, app := range apps {
			keys = append(keys, relatedresource.Key{
				Namespace: app.Namespace,
				Name:      app.Name,
			})
		}

		return keys, nil
	case *adminv1.ClusterDomain:
		return []relatedresource.Key{
			{
				Namespace: "*",
				Name:      "*",
			},
		}, nil
	}

	return nil, nil
}

func (r *Resolver) RouterEndpoints(router *riov1.Router) ([]string, error) {
	if !domains.IsPublicRouter(router) {
		return []string{
			router.Name,
			fmt.Sprintf("%s.%s", router.Name, router.Namespace),
			fmt.Sprintf("%s.%s.svc.cluster.local", router.Name, router.Namespace),
		}, nil
	}

	var endpoints []string

	pd, err := r.endpointsFromPublicDomain(router.Namespace, router.Name, "")
	if err != nil {
		return nil, err
	}
	endpoints = append(endpoints, pd...)

	name := name2.SafeConcatName(router.Name, router.Namespace)
	domains, err := r.endpointsFromClusterDomain(name)
	if err != nil {
		return nil, err
	}
	endpoints = append(endpoints, domains...)

	return endpoints, nil
}

func (r *Resolver) AppEndpoints(service *riov1.Service) ([]string, error) {
	if !domains.IsPublic(service) {
		return nil, nil
	}

	var endpoints []string

	app, _ := services.AppAndVersion(service)
	pd, err := r.endpointsFromPublicDomain(service.Namespace, app, "")
	if err != nil {
		return nil, err
	}
	endpoints = append(endpoints, pd...)

	name := name2.SafeConcatName(app, service.Namespace)
	domains, err := r.endpointsFromClusterDomain(name)
	if err != nil {
		return nil, err
	}
	endpoints = append(endpoints, domains...)

	return endpoints, nil
}

func (r *Resolver) ServiceEndpoints(service *riov1.Service) ([]string, error) {
	if !domains.IsPublic(service) {
		return nil, nil
	}

	var endpoints []string

	app, version := services.AppAndVersion(service)
	pd, err := r.endpointsFromPublicDomain(service.Namespace, app, version)
	if err != nil {
		return nil, err
	}
	endpoints = append(endpoints, pd...)

	name := name2.SafeConcatName(app, version, service.Namespace)
	domains, err := r.endpointsFromClusterDomain(name)
	if err != nil {
		return nil, err
	}
	endpoints = append(endpoints, domains...)

	return endpoints, nil
}

func (r *Resolver) endpointsFromPublicDomain(namespace, app, version string) ([]string, error) {
	var (
		endpoints []string
		key       string
	)

	if version == "" {
		key = fmt.Sprintf("%s/%s", namespace, app)
	} else {
		key = fmt.Sprintf("%s/%s/%s", namespace, app, version)
	}

	pds, err := r.publicDomainCache.GetByIndex(indexes.PublicDomainByTarget, key)
	if err != nil {
		return nil, err
	}

	sort.Slice(pds, func(i, j int) bool {
		return pds[i].Name < pds[j].Name
	})

	for _, pd := range pds {
		scheme := "http"
		if pd.Status.HTTPSSupported {
			scheme = "https"
		}
		endpoints = append(endpoints, fmt.Sprintf("%s://%s", scheme, pd.Name))
	}

	return endpoints, nil
}

func (r *Resolver) endpointsFromClusterDomain(name string) ([]string, error) {
	var endpoints []string

	cds, err := r.clusterDomainCache.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	sort.Slice(cds, func(i, j int) bool {
		return cds[i].Name < cds[j].Name
	})

	for _, cd := range cds {
		if cd.Status.HTTPSSupported {
			endpoints = append(endpoints, formatEndpoint("https", cd.Spec.HTTPSPort, 443, name, cd.Name))
		}
		endpoints = append(endpoints, formatEndpoint("http", cd.Spec.HTTPPort, 80, name, cd.Name))
	}

	return endpoints, nil
}

func formatEndpoint(scheme string, port, defaultPort int, name, hostname string) string {
	if port == 0 {
		return ""
	}
	if port == defaultPort {
		return fmt.Sprintf("%s://%s.%s", scheme, name, hostname)
	}
	return fmt.Sprintf("%s://%s.%s:%d", scheme, name, hostname, port)
}
