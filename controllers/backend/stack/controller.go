package stack

import (
	"context"
	"strings"

	"github.com/rancher/norman/pkg/changeset"
	"github.com/rancher/rio/pkg/deploy/stack"
	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	stackByNS = "stackByNS"
)

func Register(ctx context.Context, rContext *types.Context) error {
	s := &stackDeployController{
		stacks:        rContext.Rio.Stack,
		stackCache:    rContext.Rio.Stack.Cache(),
		serviceCache:  rContext.Rio.Service.Cache(),
		configCache:   rContext.Rio.Config.Cache(),
		volumeCache:   rContext.Rio.Volume.Cache(),
		routeSetCache: rContext.Rio.RouteSet.Cache(),
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
		rContext.Rio.Stack)

	rContext.Rio.Stack.Cache().Index(stackByNS, index)

	return nil
}

type stackDeployController struct {
	stacks        v1beta1.StackClient
	stackCache    v1beta1.StackClientCache
	serviceCache  v1beta1.ServiceClientCache
	configCache   v1beta1.ConfigClientCache
	volumeCache   v1beta1.VolumeClientCache
	routeSetCache v1beta1.RouteSetClientCache
}

func index(stack *v1beta1.Stack) ([]string, error) {
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

func (s *stackDeployController) Remove(obj *v1beta1.Stack) (runtime.Object, error) {
	err := stack.Remove(namespace.StackToNamespace(obj), getSpace(obj), obj)
	return obj, err
}

func (s *stackDeployController) Updated(obj *v1beta1.Stack) (runtime.Object, error) {
	// Wait until defined
	if !v1beta1.StackConditionDefined.IsTrue(obj) {
		return obj, nil
	}

	_, err := v1beta1.StackConditionDeployed.Do(obj, func() (runtime.Object, error) {
		return s.deploy(obj)
	})
	return obj, err
}

func (s *stackDeployController) deploy(obj *v1beta1.Stack) (*v1beta1.Stack, error) {
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

	err = stack.Deploy(namespace,
		getSpace(obj),
		obj,
		configs,
		services,
		volumes,
		routes)
	return obj, err
}

func getSpace(stack *v1beta1.Stack) string {
	parts := strings.SplitN(stack.Namespace, "-", 2)
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}
