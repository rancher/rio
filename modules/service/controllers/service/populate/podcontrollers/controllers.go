package podcontrollers

import (
	"github.com/rancher/rio/modules/service/controllers/service/populate/pod"
	"github.com/rancher/rio/modules/service/controllers/service/populate/servicelabels"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
)

func Populate(service *riov1.Service, os *objectset.ObjectSet) error {
	podTemplateSpec, err := pod.Populate(service, os)
	if err != nil {
		return err
	}

	if !allImagesSet(podTemplateSpec) {
		return nil
	}

	cp := newControllerParams(service, podTemplateSpec)
	if service.Spec.Global {
		daemonset(service, cp, os)
	} else if len(cp.VolumeTemplates) > 0 {
		statefulset(service, cp, os)
	} else {
		deployment(service, cp, os)
	}

	return nil
}

func allImagesSet(podTemplate v1.PodTemplateSpec) bool {
	for _, container := range podTemplate.Spec.Containers {
		if container.Image == "" {
			return false
		}
	}
	for _, container := range podTemplate.Spec.InitContainers {
		if container.Image == "" {
			return false
		}
	}
	return true
}

func newControllerParams(service *riov1.Service, podTemplateSpec v1.PodTemplateSpec) *controllerParams {
	scaleParams := parseScaleParams(service)
	selectorLabels := servicelabels.SelectorLabels(service)
	labels := servicelabels.ServiceLabels(service)
	volumeTemplates := pod.NormalizeVolumeTemplates(service)
	annotations := annotations(service)

	// Selector labels must be on the podTemplateSpec
	podTemplateSpec.Labels = servicelabels.Merge(podTemplateSpec.Labels, selectorLabels)

	return &controllerParams{
		Scale:           scaleParams,
		Labels:          labels,
		Annotations:     annotations,
		SelectorLabels:  selectorLabels,
		PodTemplateSpec: podTemplateSpec,
		VolumeTemplates: volumeTemplates,
	}
}

func annotations(service *riov1.Service) map[string]string {
	result := map[string]string{}
	if service.Spec.ServiceMesh != nil && !*service.Spec.ServiceMesh {
		result["rio.cattle.io/mesh"] = "false"
	} else {
		result["rio.cattle.io/mesh"] = "true"
	}
	return result
}

type controllerParams struct {
	Scale           scaleParams
	Labels          map[string]string
	Annotations     map[string]string
	SelectorLabels  map[string]string
	VolumeTemplates map[string]riov1.VolumeTemplate
	PodTemplateSpec v1.PodTemplateSpec
}
