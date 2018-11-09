package stack

import (
	"context"

	"github.com/rancher/norman/pkg/changeset"
	"github.com/rancher/rio/pkg/deploy/stack"
	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

const (
	stackByNS = "stackByNS"
)

func Register(ctx context.Context, rContext *types.Context) error {
	s := &stackDeployController{
		stacks:             rContext.Rio.Stacks(""),
		stackController:    rContext.Rio.Stacks("").Controller(),
		serviceController:  rContext.Rio.Services("").Controller(),
		configController:   rContext.Rio.Configs("").Controller(),
		volumeController:   rContext.Rio.Volumes("").Controller(),
		routeSetController: rContext.Rio.RouteSets("").Controller(),
	}

	rContext.Rio.Stacks("").AddLifecycle(ctx, "stack-deploy-controller", s)
	changeset.Watch(ctx, "stack-deploy",
		s.resolve,
		s.stackController.Enqueue,
		s.serviceController,
		s.configController,
		s.volumeController,
		s.routeSetController,
		s.stackController)

	s.stackController.Informer().AddIndexers(cache.Indexers{
		stackByNS: index,
	})

	return nil
}

type stackDeployController struct {
	stacks             v1beta1.StackInterface
	stackController    v1beta1.StackController
	serviceController  v1beta1.ServiceController
	configController   v1beta1.ConfigController
	volumeController   v1beta1.VolumeController
	routeSetController v1beta1.RouteSetController
}

func index(obj interface{}) ([]string, error) {
	stack, ok := obj.(*v1beta1.Stack)
	if !ok || stack == nil {
		return nil, nil
	}
	return []string{
		namespace.StackToNamespace(stack),
	}, nil
}

func (s *stackDeployController) resolve(ns, name string, obj runtime.Object) ([]changeset.Key, error) {
	objs, err := s.stackController.Informer().GetIndexer().ByIndex(stackByNS, ns)
	if err != nil {
		return nil, nil
	}

	if len(objs) != 1 {
		return nil, nil
	}

	stack, ok := objs[0].(*v1beta1.Stack)
	if !ok {
		return nil, nil
	}

	return []changeset.Key{
		{
			Namespace: stack.Namespace,
			Name:      stack.Name,
		},
	}, nil
}

func (s *stackDeployController) Create(obj *v1beta1.Stack) (runtime.Object, error) {
	return nil, nil
}

func (s *stackDeployController) Remove(obj *v1beta1.Stack) (runtime.Object, error) {
	err := stack.Remove(namespace.StackToNamespace(obj), obj)
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

	configs, err := s.configController.Lister().List(namespace, labels.Everything())
	if err != nil {
		return nil, err
	}

	volumes, err := s.volumeController.Lister().List(namespace, labels.Everything())
	if err != nil {
		return nil, err
	}

	services, err := s.serviceController.Lister().List(namespace, labels.Everything())
	if err != nil {
		return nil, err
	}

	routes, err := s.routeSetController.Lister().List(namespace, labels.Everything())
	if err != nil {
		return nil, err
	}

	err = stack.Deploy(namespace,
		obj,
		configs,
		services,
		volumes,
		routes)
	return obj, err
}
