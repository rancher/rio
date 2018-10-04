package parse

import (
	"github.com/rancher/rancher/pkg/ref"
	"github.com/rancher/rio/pkg/deploy/stackdef/output"
	"github.com/rancher/rio/pkg/deploy/stackdef/populate/crd"
	"github.com/rancher/rio/pkg/deploy/stackdef/populate/k8s"
	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/pkg/template"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
)

func Populate(stack *v1beta1.Stack, output *output.Deployment) error {
	internalStack, err := parseStack(stack)
	if err != nil {
		return err
	}

	ns := namespace.StackToNamespace(stack)
	output.Namespace = ns

	if stack.Spec.EnableKubernetesResources {
		if err := crd.Populate(internalStack, output); err != nil {
			return err
		}
		if err := k8s.Populate(internalStack, output); err != nil {
			return err
		}
	}

	configs(ns, stack, internalStack, output)
	volumes(ns, stack, internalStack, output)
	services(ns, stack, internalStack, output)

	return nil
}

func parseStack(stack *v1beta1.Stack) (*v1beta1.InternalStack, error) {
	t, err := template.FromStack(stack)
	if err != nil {
		return nil, err
	}

	if err := t.Validate(); err != nil {
		return nil, err
	}

	return t.ToInternalStack()
}

func configs(ns string, stack *v1beta1.Stack, internalStack *v1beta1.InternalStack, output *output.Deployment) {
	for name, config := range internalStack.Configs {
		newResource := config.DeepCopy()
		newResource.Kind = "Config"
		newResource.APIVersion = v1beta1.SchemeGroupVersion.String()
		newResource.Name = name
		newResource.Namespace = ns
		newResource.Spec.SpaceName = stack.Namespace
		newResource.Spec.StackName = ref.FromStrings(stack.Namespace, stack.Name)

		output.Configs[newResource.Name] = newResource
	}
}

func volumes(ns string, stack *v1beta1.Stack, internalStack *v1beta1.InternalStack, output *output.Deployment) {
	for name, volume := range internalStack.Volumes {
		newResource := volume.DeepCopy()
		newResource.Kind = "Volume"
		newResource.APIVersion = v1beta1.SchemeGroupVersion.String()
		newResource.Name = name
		newResource.Namespace = ns
		newResource.Spec.SpaceName = stack.Namespace
		newResource.Spec.StackName = ref.FromStrings(stack.Namespace, stack.Name)

		output.Volumes[newResource.Name] = newResource
	}
}

func services(ns string, stack *v1beta1.Stack, internalStack *v1beta1.InternalStack, output *output.Deployment) {
	for name, service := range internalStack.Services {
		newResource := service.DeepCopy()
		newResource.Kind = "Service"
		newResource.APIVersion = v1beta1.SchemeGroupVersion.String()
		newResource.Name = name
		newResource.Namespace = ns
		newResource.Spec.SpaceName = stack.Namespace
		newResource.Spec.StackName = ref.FromStrings(stack.Namespace, stack.Name)
		newResource.Status.Conditions = []v1beta1.Condition{
			{
				Type:   "Pending",
				Status: "Unknown",
			},
		}

		output.Services[newResource.Name] = newResource
	}
}
