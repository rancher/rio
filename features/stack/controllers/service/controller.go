package service

import (
	"context"

	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/features/routing/pkg/istio/config"
	"github.com/rancher/rio/features/stack/controllers/service/populate"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	cf := config.NewConfigFactory(ctx, rContext.Core.ConfigMap.Interface(),
		settings.IstioExternalLBNamespace,
		settings.IstionConfigMapName,
		settings.IstionConfigMapKey)
	injector := config.NewIstioInjector(cf)

	c := stackobject.NewGeneratingController(ctx, rContext, "stack-service", rContext.Rio.Service, *injector)
	c.Processor.Client(
		rContext.RBAC.Role,
		rContext.RBAC.RoleBinding,
		rContext.RBAC.ClusterRole,
		rContext.RBAC.ClusterRoleBinding,
		rContext.Apps.DaemonSet,
		rContext.Apps.Deployment,
		rContext.Apps.StatefulSet,
		rContext.Policy.PodDisruptionBudget,
		rContext.Core.ServiceAccount,
		rContext.Core.Service)

	sh := &serviceHandler{
		serviceClient: rContext.Rio.Service,
		serviceCache:  rContext.Rio.Service.Cache(),
		configCache:   rContext.Rio.Config.Cache(),
		volumeCache:   rContext.Rio.Volume.Cache(),
	}

	c.Populator = sh.populate
	rContext.Rio.Service.OnChange(ctx, "stack-service-change-controller", sh.onChange)

	return nil
}

type serviceHandler struct {
	serviceClient riov1.ServiceClient
	serviceCache  riov1.ServiceClientCache
	configCache   riov1.ConfigClientCache
	volumeCache   riov1.VolumeClientCache
}

func (s *serviceHandler) onChange(service *riov1.Service) (runtime.Object, error) {
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

	return populate.Service(stack, configsByName, volumesByName, services, service, os)
}
