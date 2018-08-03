package stack

import (
	"github.com/rancher/rancher/pkg/ref"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (s *stackController) gatherObjects(ns string, stack *v1beta1.Stack, internalStack *v1beta1.InternalStack) []runtime.Object {
	var resources []runtime.Object

	for name, config := range internalStack.Configs {
		newResource := config.DeepCopy()
		newResource.Kind = "Config"
		newResource.APIVersion = v1beta1.SchemeGroupVersion.String()
		newResource.Name = name
		newResource.Namespace = ns
		newResource.Spec.SpaceName = stack.Namespace
		newResource.Spec.StackName = ref.FromStrings(stack.Namespace, stack.Name)

		resources = append(resources, newResource)
	}

	for name, volume := range internalStack.Volumes {
		newResource := volume.DeepCopy()
		newResource.Kind = "Volume"
		newResource.APIVersion = v1beta1.SchemeGroupVersion.String()
		newResource.Name = name
		newResource.Namespace = ns
		newResource.Spec.SpaceName = stack.Namespace
		newResource.Spec.StackName = ref.FromStrings(stack.Namespace, stack.Name)

		resources = append(resources, newResource)
	}

	for name, service := range internalStack.Services {
		newResource := service.DeepCopy()
		newResource.Kind = "Service"
		newResource.APIVersion = v1beta1.SchemeGroupVersion.String()
		newResource.Name = name
		newResource.Namespace = ns
		newResource.Spec.SpaceName = stack.Namespace
		newResource.Spec.StackName = ref.FromStrings(stack.Namespace, stack.Name)

		resources = append(resources, newResource)
	}

	return resources
}
