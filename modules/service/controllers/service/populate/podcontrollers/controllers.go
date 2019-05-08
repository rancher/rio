package podcontrollers

import (
	"github.com/rancher/rio/modules/service/controllers/service/populate/pod"
	"github.com/rancher/rio/modules/service/controllers/service/populate/servicelabels"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
)

func Populate(service *riov1.Service, os *objectset.ObjectSet) error {
	if service.Spec.SystemSpec != nil {
		cp := newControllerParams(service, v1.PodTemplateSpec{Spec: service.Spec.SystemSpec.PodSpec})
		deployment(service, cp, os)
		return nil
	}

	if !isImageSet(service) {
		return nil
	}

	podTemplateSpec, err := pod.Populate(service, os)
	if err != nil {
		return err
	}

	cp := newControllerParams(service, podTemplateSpec)
	deployment(service, cp, os)

	return nil
}

func isImageSet(service *riov1.Service) bool {
	if service.Spec.Image == "" && service.Spec.Build != nil {
		return false
	}
	for _, con := range service.Spec.Sidecars {
		if con.Image == "" && con.Build != nil {
			return false
		}
	}
	return true
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

	if service.Status.ObservedScale != nil {
		scaleParams.Scale = int32(*service.Status.ObservedScale)
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
