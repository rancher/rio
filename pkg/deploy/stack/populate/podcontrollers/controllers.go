package podcontrollers

import (
	"github.com/rancher/rio/pkg/deploy/stack/input"
	"github.com/rancher/rio/pkg/deploy/stack/output"
	"github.com/rancher/rio/pkg/deploy/stack/populate/pod"
	"github.com/rancher/rio/pkg/deploy/stack/populate/podvolume"
	"github.com/rancher/rio/pkg/deploy/stack/populate/servicelabels"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"k8s.io/api/core/v1"
)

func Populate(stack *input.Stack, service *v1beta1.Service, output *output.Deployment) error {
	podTemplateSpec := pod.Populate(stack, service, output)

	cp := newControllerParams(stack, service, podTemplateSpec)
	usedTemplates := podvolume.UsedTemplates(stack, service)

	pdb(stack, service, cp, output)

	if service.Spec.Global {
		daemonSet(stack, service, cp, output)
	} else if isDeployment(service.Spec.ServiceUnversionedSpec, usedTemplates) {
		deployment(stack, service, cp, output)
	} else {
		return statefulSet(stack, service, cp, usedTemplates, output)
	}

	return nil
}

func newControllerParams(stack *input.Stack, service *v1beta1.Service, podTemplateSpec v1.PodTemplateSpec) *controllerParams {
	scaleParams := parseScaleParams(&service.Spec.ServiceUnversionedSpec)
	selectorLabels := servicelabels.SelectorLabels(stack, service)
	labels := servicelabels.ServiceLabels(stack, service)

	if podTemplateSpec.Labels == nil {
		podTemplateSpec.Labels = map[string]string{}
	}
	for k, v := range selectorLabels {
		podTemplateSpec.Labels[k] = v
	}

	return &controllerParams{
		Scale:           scaleParams,
		Labels:          labels,
		SelectorLabels:  selectorLabels,
		PodTemplateSpec: podTemplateSpec,
	}
}

type controllerParams struct {
	Scale           scaleParams
	Labels          map[string]string
	SelectorLabels  map[string]string
	PodTemplateSpec v1.PodTemplateSpec
}
