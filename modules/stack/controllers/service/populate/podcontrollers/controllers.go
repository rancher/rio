package podcontrollers

import (
	"github.com/rancher/rio/modules/stack/controllers/service/populate/pod"
	"github.com/rancher/rio/modules/stack/controllers/service/populate/podvolume"
	"github.com/rancher/rio/modules/stack/controllers/service/populate/servicelabels"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
)

func Populate(configsByName map[string]*riov1.Config, volumesByName map[string]*riov1.Volume, service *riov1.Service, os *objectset.ObjectSet) error {
	podTemplateSpec := pod.Populate(configsByName, volumesByName, service, os)

	cp := newControllerParams(service, podTemplateSpec)
	usedTemplates := podvolume.UsedTemplates(volumesByName, service)

	pdb(service, cp, os)

	if service.Spec.Global {
		daemonSet(service, cp, os)
	} else if isDeployment(service.Spec.ServiceUnversionedSpec, usedTemplates) {
		deployment(service, cp, os)
	} else {
		return statefulSet(service, cp, usedTemplates, os)
	}

	return nil
}

func newControllerParams(service *riov1.Service, podTemplateSpec v1.PodTemplateSpec) *controllerParams {
	scaleParams := parseScaleParams(&service.Spec.ServiceUnversionedSpec)
	selectorLabels := servicelabels.SelectorLabels(service)
	labels := servicelabels.ServiceLabels(service)

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
