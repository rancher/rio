package globalrbac

import (
	"context"
	"fmt"

	"github.com/rancher/rio/modules/service/controllers/service/populate/rbac"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	rbacv1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/rbac/v1"
	"github.com/rancher/wrangler/pkg/generic"
	"github.com/rancher/wrangler/pkg/name"
	"github.com/rancher/wrangler/pkg/objectset"
	"github.com/rancher/wrangler/pkg/relatedresource"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	handlerName = "service-cluster-rbac"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := &handler{
		serviceCache: rContext.Rio.Rio().V1().Service().Cache(),
		crbClient:    rContext.RBAC.Rbac().V1().ClusterRoleBinding(),
		crClient:     rContext.RBAC.Rbac().V1().ClusterRole(),
	}

	riov1controller.RegisterServiceGeneratingHandler(ctx,
		rContext.Rio.Rio().V1().Service(),
		rContext.Apply.
			WithStrictCaching().
			WithCacheTypes(rContext.RBAC.Rbac().V1().ClusterRole(),
				rContext.RBAC.Rbac().V1().ClusterRoleBinding()).
			WithSetID(handlerName),
		"ServiceClusterRBAC",
		handlerName,
		h.generate,
		&generic.GeneratingHandlerOptions{
			AllowClusterScoped: true,
		})

	rContext.RBAC.Rbac().V1().ClusterRoleBinding().OnChange(ctx, handlerName, h.onClusterRoleBinding)
	rContext.RBAC.Rbac().V1().ClusterRole().OnChange(ctx, handlerName, h.onClusterRole)

	// watch the rio service for changes in order to propagate changes to the ClusterRole controller
	clusterBindWrap := ClusterScopedRoleBindingWrappper{rContext.RBAC.Rbac().V1().ClusterRoleBinding()}
	clusterRoleWrap := ClusterScopedRoleWrapper{rContext.RBAC.Rbac().V1().ClusterRole()}

	relatedresource.Watch(ctx, handlerName, resolveToCR, clusterRoleWrap, rContext.Rio.Rio().V1().Service())
	relatedresource.Watch(ctx, handlerName, resolveToCRBinding, clusterBindWrap, rContext.Rio.Rio().V1().Service())

	return nil
}

type ClusterScopedRoleWrapper struct {
	actual rbacv1controller.ClusterRoleController
}

type ClusterScopedRoleBindingWrappper struct {
	actual rbacv1controller.ClusterRoleBindingController
}

func (c ClusterScopedRoleWrapper) Enqueue(namespace string, name string) {
	c.actual.Enqueue(name)
}

func (c ClusterScopedRoleBindingWrappper) Enqueue(namespace string, name string) {
	c.actual.Enqueue(name)

}

// resolver is going to run whenever Rio Service is updated (obj is Rio Service), should only enqueue changes when a service is modified/deleted
// need to return a list of keys to cluster roles that are associated with that service
func resolveToCR(namespace, n string, obj runtime.Object) ([]relatedresource.Key, error) {
	if obj == nil {
		return nil, nil
	}
	serviceName, serviceNamespace, err := serviceNameAndNamespace(obj)
	if err != nil {
		return nil, err
	}
	// cluster role names follow a format of
	// rio-$namespace-$serviceName
	clusterRole := name.SafeConcatName("rio", serviceNamespace, serviceName)
	return []relatedresource.Key{
		{
			Namespace: "",
			Name:      clusterRole,
		},
	}, nil
}

// resolver is going to run whenever Rio Service is updated (obj is Rio Service), should only enqueue changes when a service is modified/deleted
// need to return a list of keys to cluster roles that are associated with that service
func resolveToCRBinding(namespace, n string, obj runtime.Object) ([]relatedresource.Key, error) {
	if obj == nil {
		return nil, nil
	}
	serviceName, serviceNamespace, err := serviceNameAndNamespace(obj)
	if err != nil {
		return nil, err
	}
	// cluster role binding follow a format of
	// rio-$namespace-$serviceName-rio-$namespace-$serviceName
	clusterBinding := name.SafeConcatName("rio", serviceNamespace, serviceName, "rio", serviceNamespace, serviceName)
	return []relatedresource.Key{
		{
			Namespace: "",
			Name:      clusterBinding,
		},
	}, nil
}

type handler struct {
	serviceCache riov1controller.ServiceCache
	crbClient    rbacv1controller.ClusterRoleBindingClient
	crClient     rbacv1controller.ClusterRoleClient
}

func (h *handler) onClusterRoleBinding(key string, crb *rbacv1.ClusterRoleBinding) (*rbacv1.ClusterRoleBinding, error) {
	if crb == nil || crb.DeletionTimestamp != nil {
		return nil, nil
	}

	svcName := crb.Labels["rio.cattle.io/service"]
	ns := crb.Labels["rio.cattle.io/namespace"]

	if svcName == "" || ns == "" {
		return crb, nil
	}

	_, err := h.serviceCache.Get(ns, svcName)

	if errors.IsNotFound(err) {
		return crb, h.crbClient.Delete(crb.Name, nil)
	}

	return crb, nil
}

func (h *handler) onClusterRole(key string, cr *rbacv1.ClusterRole) (*rbacv1.ClusterRole, error) {
	if cr == nil || cr.DeletionTimestamp != nil {
		return nil, nil
	}

	svcName := cr.Labels["rio.cattle.io/service"]
	ns := cr.Labels["rio.cattle.io/namespace"]

	if svcName == "" || ns == "" {
		return cr, nil
	}

	_, err := h.serviceCache.Get(ns, svcName)
	if errors.IsNotFound(err) {
		return cr, h.crClient.Delete(cr.Name, nil)
	}

	return cr, nil
}

func (h *handler) generate(obj *riov1.Service, status riov1.ServiceStatus) ([]runtime.Object, riov1.ServiceStatus, error) {
	os := objectset.NewObjectSet()
	if err := rbac.PopulateCluster(obj, os); err != nil {
		return nil, status, err
	}
	return os.All(), status, nil
}

func serviceNameAndNamespace(obj runtime.Object) (serviceName string, serviceNamespace string, err error) {
	service, ok := obj.(*riov1.Service)
	if !ok {
		return "", "", fmt.Errorf("type assertion failed; obj -> rio service")
	}
	return service.Name, service.Namespace, nil
}
