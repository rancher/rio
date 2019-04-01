package service

import (
	"context"

	"github.com/rancher/rio/modules/stack/controllers/service/populate"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/generated/controllers/build.knative.dev/v1alpha1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	buildkitAddr = "buildkit.build"
)

func Register(ctx context.Context, rContext *types.Context) error {
	//cf := config.NewConfigFactory(ctx, rContext.Core.Core().V1().ConfigMap(),
	//	settings.IstioStackName,
	//	settings.MeshConfigMapName,
	//	settings.IstionConfigMapKey)
	//injector := config.NewIstioInjector(cf)

	c := stackobject.NewGeneratingController(ctx, rContext, "stack-service", rContext.Rio.Rio().V1().Service(), "istio")
	c.Apply = c.Apply.WithCacheTypes(
		rContext.Build.Build().V1alpha1().Build(),
		rContext.RBAC.Rbac().V1().Role(),
		rContext.RBAC.Rbac().V1().RoleBinding(),
		rContext.RBAC.Rbac().V1().ClusterRole(),
		rContext.RBAC.Rbac().V1().ClusterRoleBinding(),
		rContext.Apps.Apps().V1().DaemonSet(),
		rContext.Apps.Apps().V1().Deployment(),
		rContext.Apps.Apps().V1().StatefulSet(),
		rContext.Policy.Policy().V1beta1().PodDisruptionBudget(),
		rContext.Core.Core().V1().ServiceAccount(),
		rContext.Core.Core().V1().Service(),
		rContext.Core.Core().V1().Secret(),
		rContext.AutoScale.Autoscale().V1().ServiceScaleRecommendation(),
		rContext.Webhook.Webhookinator().V1().GitWebHookReceiver())

	sh := &serviceHandler{
		namespace:     rContext.Namespace,
		serviceClient: rContext.Rio.Rio().V1().Service(),
		serviceCache:  rContext.Rio.Rio().V1().Service().Cache(),
		configCache:   rContext.Rio.Rio().V1().Config().Cache(),
		volumeCache:   rContext.Rio.Rio().V1().Volume().Cache(),
		buildCache:    rContext.Build.Build().V1alpha1().Build().Cache(),
	}

	c.Populator = sh.populate
	rContext.Rio.Rio().V1().Service().OnChange(ctx, "stack-service-change-controller", sh.onChange)

	return nil
}

type serviceHandler struct {
	namespace     string
	serviceClient riov1controller.ServiceController
	serviceCache  riov1controller.ServiceCache
	configCache   riov1controller.ConfigCache
	volumeCache   riov1controller.VolumeCache
	buildCache    v1alpha1.BuildCache
}

func (s *serviceHandler) onChange(key string, service *riov1.Service) (*riov1.Service, error) {
	if service == nil {
		return nil, nil
	}

	if service.Spec.Revision.ParentService != "" {
		// enqueue parent so that we re-evaluate the destionationRules
		s.serviceClient.Enqueue(service.Namespace, service.Spec.Revision.ParentService)
	}

	return service, nil
}

func (s *serviceHandler) populate(obj runtime.Object, stack *riov1.Stack, os *objectset.ObjectSet) error {
	service := obj.(*riov1.Service)
	services, err := s.serviceCache.List(service.Namespace, labels.Everything())
	if err != nil {
		return err
	}

	configsByName := map[string]*riov1.Config{}
	configs, err := s.configCache.List(service.Namespace, labels.Everything())
	if err != nil {
		return err
	}
	for _, config := range configs {
		configsByName[config.Name] = config
	}

	volumesByName := map[string]*riov1.Volume{}
	volumes, err := s.volumeCache.List(service.Namespace, labels.Everything())
	if err != nil {
		return err
	}
	for _, volume := range volumes {
		volumesByName[volume.Name] = volume
	}

	return populate.Service(configsByName, volumesByName, services, service, os)
}
