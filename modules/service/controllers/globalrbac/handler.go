package globalrbac

import (
	"context"

	"github.com/rancher/rio/modules/service/controllers/service/populate/rbac"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	rbacv1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/rbac/v1"
	"github.com/rancher/wrangler/pkg/generic"
	"github.com/rancher/wrangler/pkg/objectset"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
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

	return nil
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
