package parse

import (
	"github.com/rancher/rio/modules/stack/controllers/stack/populate/crd"
	"github.com/rancher/rio/modules/stack/controllers/stack/populate/k8s"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackfile"
	"github.com/rancher/wrangler/pkg/objectset"
)

func Populate(stack *riov1.Stack, output *objectset.ObjectSet) error {
	internalStack, err := parseStack(stack)
	if err != nil {
		return err
	}

	ns := stack.Name

	if stack.Spec.EnableKubernetesResources {
		if err := crd.Populate(internalStack, output); err != nil {
			return err
		}
		if err := k8s.Populate(stack, internalStack, output); err != nil {
			return err
		}
	}

	configs(ns, stack, internalStack, output)
	volumes(ns, stack, internalStack, output)
	services(ns, stack, internalStack, output)
	routes(ns, stack, internalStack, output)
	externalservices(ns, stack, internalStack, output)

	return nil
}

func parseStack(stack *riov1.Stack) (*riov1.StackFile, error) {
	t, err := stackfile.FromStack(stack)
	if err != nil {
		return nil, err
	}

	if err := t.Validate(); err != nil {
		return nil, err
	}

	return t.ToStackResources()
}

func configs(ns string, stack *riov1.Stack, internalStack *riov1.StackFile, output *objectset.ObjectSet) {
	for name, config := range internalStack.Configs {
		newResource := config.DeepCopy()
		newResource.Kind = "Config"
		newResource.APIVersion = riov1.SchemeGroupVersion.String()
		newResource.Name = name
		newResource.Namespace = ns

		output.Add(newResource)
	}
}

func volumes(ns string, stack *riov1.Stack, internalStack *riov1.StackFile, output *objectset.ObjectSet) {
	for name, volume := range internalStack.Volumes {
		newResource := volume.DeepCopy()
		newResource.Kind = "Volume"
		newResource.APIVersion = riov1.SchemeGroupVersion.String()
		newResource.Name = name
		newResource.Namespace = ns

		output.Add(newResource)
	}
}

func services(ns string, stack *riov1.Stack, internalStack *riov1.StackFile, output *objectset.ObjectSet) {
	for name, service := range internalStack.Services {
		newResource := service.DeepCopy()
		newResource.Kind = "Service"
		newResource.APIVersion = riov1.SchemeGroupVersion.String()
		newResource.Name = name
		newResource.Namespace = ns

		output.Add(newResource)
	}
}

func routes(ns string, stack *riov1.Stack, internalStack *riov1.StackFile, output *objectset.ObjectSet) {
	for name, routes := range internalStack.Routes {
		newResource := routes.DeepCopy()
		newResource.Kind = "RouteSet"
		newResource.APIVersion = riov1.SchemeGroupVersion.String()
		newResource.Name = name
		newResource.Namespace = ns

		output.Add(newResource)
	}
}

func externalservices(ns string, stack *riov1.Stack, internalStack *riov1.StackFile, output *objectset.ObjectSet) {
	for name, routes := range internalStack.ExternalServices {
		newResource := routes.DeepCopy()
		newResource.Kind = "ExternalService"
		newResource.APIVersion = riov1.SchemeGroupVersion.String()
		newResource.Name = name
		newResource.Namespace = ns

		output.Add(newResource)
	}
}
