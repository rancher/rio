package service

import (
	"context"

	"github.com/rancher/rio/modules/service/controllers/service/populate"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/objectset"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	scs, err := rContext.Storage.Storage().V1().StorageClass().List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, sc := range scs.Items {
		if sc.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
			constants.DefaultStorageClass = true
			break
		}
	}

	c := stackobject.NewGeneratingController(ctx, rContext, "stack-service", rContext.Rio.Rio().V1().Service())
	c.Apply = c.Apply.WithCacheTypes(
		rContext.Build.Tekton().V1alpha1().TaskRun(),
		rContext.RBAC.Rbac().V1().Role(),
		rContext.RBAC.Rbac().V1().RoleBinding(),
		rContext.RBAC.Rbac().V1().ClusterRole(),
		rContext.RBAC.Rbac().V1().ClusterRoleBinding(),
		rContext.Apps.Apps().V1().Deployment(),
		rContext.Apps.Apps().V1().DaemonSet(),
		rContext.Core.Core().V1().ServiceAccount(),
		rContext.Core.Core().V1().Service(),
		rContext.Core.Core().V1().Secret(),
		rContext.Core.Core().V1().PersistentVolumeClaim(),
		rContext.AutoScale.Autoscale().V1().ServiceScaleRecommendation(),
		rContext.Webhook.Gitwatcher().V1().GitWatcher()).
		WithRateLimiting(5).
		WithStrictCaching()

	sh := &serviceHandler{
		namespace:     rContext.Namespace,
		serviceClient: rContext.Rio.Rio().V1().Service(),
		serviceCache:  rContext.Rio.Rio().V1().Service().Cache(),
		ns:            rContext.Core.Core().V1().Namespace(),
	}

	c.Populator = sh.populate
	return nil
}

type serviceHandler struct {
	namespace     string
	serviceClient riov1controller.ServiceController
	serviceCache  riov1controller.ServiceCache
	ns            corev1controller.NamespaceController
}

func (s *serviceHandler) populate(obj runtime.Object, ns *corev1.Namespace, os *objectset.ObjectSet) error {
	service := obj.(*riov1.Service)

	ns, err := s.ns.Cache().Get(service.Namespace)
	if err != nil {
		return err
	}

	if ns.Name != s.namespace && constants.ServiceMeshMode == constants.ServiceMeshModeIstio {
		ns = ns.DeepCopy()
		if ns.Labels == nil {
			ns.Labels = map[string]string{}
		}
		ns.Labels["istio-injection"] = "enabled"
		if _, err := s.ns.Update(ns); err != nil {
			return err
		}
	}

	if service.Namespace != s.namespace && service.SystemSpec != nil {
		service = service.DeepCopy()
		service.SystemSpec = nil
	}
	return populate.Service(service, s.namespace, os)
}
