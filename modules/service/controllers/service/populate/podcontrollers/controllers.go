package podcontrollers

import (
	"github.com/rancher/rio/modules/service/controllers/service/populate/pod"
	"github.com/rancher/rio/modules/service/controllers/service/populate/servicelabels"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
)

func Populate(service *riov1.Service, os *objectset.ObjectSet) error {
	podTemplateSpec := pod.Populate(service, os)

	cp := newControllerParams(service, podTemplateSpec)
	pdb(service, cp, os)

	if service.Spec.Global {
		daemonSet(service, cp, os)
	} else if isDeployment(service.Spec) {
		deployment(service, cp, os)
	} else {
		return statefulSet(service, cp, os)
	}

	return nil
}

func newControllerParams(service *riov1.Service, podTemplateSpec v1.PodTemplateSpec) *controllerParams {
	scaleParams := parseScaleParams(&service.Spec)
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
