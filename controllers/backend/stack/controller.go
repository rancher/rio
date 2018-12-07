package stack

import (
	"context"
	"strings"

	"github.com/rancher/norman/pkg/changeset"
	"github.com/rancher/rio/pkg/deploy/stack"
	"github.com/rancher/rio/pkg/istio/config"
	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"github.com/rancher/types/apis/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	stackByNS = "stackByNS"
)

func Register(ctx context.Context, rContext *types.Context) error {
	cf := config.NewConfigFactory(ctx, rContext.Core.ConfigMap.Interface(),
		settings.IstioExternalLBNamespace,
		settings.IstionConfigMapName,
		settings.IstionConfigMapKey)
	injector := config.NewIstioInjector(cf)
	s := &stackDeployController{
		stacks:               rContext.Rio.Stack,
		stackCache:           rContext.Rio.Stack.Cache(),
		serviceCache:         rContext.Rio.Service.Cache(),
		configCache:          rContext.Rio.Config.Cache(),
		volumeCache:          rContext.Rio.Volume.Cache(),
		routeSetCache:        rContext.Rio.RouteSet.Cache(),
		secretsCache:         rContext.Core.Secret.Cache(),
		externalServiceCache: rContext.Rio.ExternalService.Cache(),
		injector:             injector,
	}

	rContext.Rio.Stack.OnChange(ctx, "stack-deploy-controller", s.Updated)
	rContext.Rio.Stack.OnRemove(ctx, "stack-deploy-controller", s.Remove)
	changeset.Watch(ctx, "stack-deploy",
		s.resolve,
		rContext.Rio.Stack,
		rContext.Rio.Service,
		rContext.Core.ConfigMap,
		rContext.Rio.Volume,
		rContext.Rio.RouteSet,
		rContext.Rio.Stack,
		rContext.Rio.ExternalService)

	rContext.Rio.Stack.Cache().Index(stackByNS, index)

	return nil
}

type stackDeployController struct {
	stacks               riov1.StackClient
	stackCache           riov1.StackClientCache
	serviceCache         riov1.ServiceClientCache
	configCache          riov1.ConfigClientCache
	volumeCache          riov1.VolumeClientCache
	routeSetCache        riov1.RouteSetClientCache
	secretsCache         v1.SecretClientCache
	externalServiceCache riov1.ExternalServiceClientCache
	injector             *config.IstioInjector
}

func index(stack *riov1.Stack) ([]string, error) {
	return []string{
		namespace.StackToNamespace(stack),
	}, nil
}

func (s *stackDeployController) resolve(ns, name string, obj runtime.Object) ([]changeset.Key, error) {
	objs, err := s.stackCache.GetIndexed(stackByNS, ns)
	if err != nil {
		return nil, nil
	}

	if len(objs) != 1 {
		return nil, nil
	}

	stack := objs[0]
	return []changeset.Key{
		{
			Namespace: stack.Namespace,
			Name:      stack.Name,
		},
	}, nil
}

func (s *stackDeployController) Remove(obj *riov1.Stack) (runtime.Object, error) {
	err := stack.Remove(namespace.StackToNamespace(obj), getProject(obj), obj, s.injector)
	return obj, err
}

func (s *stackDeployController) Updated(obj *riov1.Stack) (runtime.Object, error) {
	// Wait until defined
	if !riov1.StackConditionDefined.IsTrue(obj) {
		return obj, nil
	}

	_, err := riov1.StackConditionDeployed.Do(obj, func() (runtime.Object, error) {
		return s.deploy(obj)
	})
	return obj, err
}

func (s *stackDeployController) deploy(obj *riov1.Stack) (*riov1.Stack, error) {
	namespace := namespace.StackToNamespace(obj)

	configs, err := s.configCache.List(namespace, labels.Everything())
	if err != nil {
		return nil, err
	}

	volumes, err := s.volumeCache.List(namespace, labels.Everything())
	if err != nil {
		return nil, err
	}

	services, err := s.serviceCache.List(namespace, labels.Everything())
	if err != nil {
		return nil, err
	}

	routes, err := s.routeSetCache.List(namespace, labels.Everything())
	if err != nil {
		return nil, err
	}

	externalServices, err := s.externalServiceCache.List(namespace, labels.Everything())
	if err != nil {
		return nil, err
	}

	err = stack.Deploy(namespace,
		getProject(obj),
		obj,
		configs,
		services,
		volumes,
		routes,
		externalServices,
		s.injector)
	return obj, err
}

func getProject(stack *riov1.Stack) string {
	parts := strings.SplitN(stack.Namespace, "-", 2)
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}
