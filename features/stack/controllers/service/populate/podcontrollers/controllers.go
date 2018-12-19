package podcontrollers

import (
	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/features/stack/controllers/service/populate/pod"
	"github.com/rancher/rio/features/stack/controllers/service/populate/podvolume"
	"github.com/rancher/rio/features/stack/controllers/service/populate/servicelabels"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"k8s.io/api/core/v1"
)

func Populate(stack *riov1.Stack, configsByName map[string]*riov1.Config, volumesByName map[string]*riov1.Volume, service *riov1.Service, os *objectset.ObjectSet) error {
	podTemplateSpec := pod.Populate(stack, configsByName, volumesByName, service, os)

	cp := newControllerParams(stack, service, podTemplateSpec)
	usedTemplates := podvolume.UsedTemplates(volumesByName, service)

	pdb(stack, service, cp, os)

	if service.Spec.Global {
		daemonSet(stack, service, cp, os)
	} else if isDeployment(service.Spec.ServiceUnversionedSpec, usedTemplates) {
		deployment(stack, service, cp, os)
	} else {
		return statefulSet(stack, service, cp, usedTemplates, os)
	}

	return nil
}

func newControllerParams(stack *riov1.Stack, service *riov1.Service, podTemplateSpec v1.PodTemplateSpec) *controllerParams {
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
