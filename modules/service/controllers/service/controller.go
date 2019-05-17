package service

import (
	"context"

	"github.com/rancher/rio/modules/service/controllers/service/populate"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "stack-service", rContext.Rio.Rio().V1().Service(), "istio-injecter")
	c.Apply = c.Apply.WithCacheTypes(
		rContext.Build.Build().V1alpha1().Build(),
		rContext.RBAC.Rbac().V1().Role(),
		rContext.RBAC.Rbac().V1().RoleBinding(),
		rContext.RBAC.Rbac().V1().ClusterRole(),
		rContext.RBAC.Rbac().V1().ClusterRoleBinding(),
		rContext.Apps.Apps().V1().Deployment(),
		rContext.Apps.Apps().V1().DaemonSet(),
		rContext.Core.Core().V1().ServiceAccount(),
		rContext.Core.Core().V1().Service(),
		rContext.Core.Core().V1().Secret(),
		rContext.AutoScale.Autoscale().V1().ServiceScaleRecommendation(),
		rContext.Webhook.Gitwatcher().V1().GitWatcher()).
		WithRateLimiting(5).
		WithStrictCaching()

	sh := &serviceHandler{
		namespace:     rContext.Namespace,
		serviceClient: rContext.Rio.Rio().V1().Service(),
		serviceCache:  rContext.Rio.Rio().V1().Service().Cache(),
	}

	c.Populator = sh.populate
	return nil
}

type serviceHandler struct {
	namespace     string
	serviceClient riov1controller.ServiceController
	serviceCache  riov1controller.ServiceCache
}

func (s *serviceHandler) populate(obj runtime.Object, ns *corev1.Namespace, os *objectset.ObjectSet) error {
	service := obj.(*riov1.Service)
	return populate.Service(service, s.namespace, os)
}
