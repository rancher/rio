package podcontrollers

import (
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type scaleParams struct {
	Scale          *int32
	MaxSurge       *intstr.IntOrString
	MaxUnavailable *intstr.IntOrString
}

func parseScaleParams(service *riov1.Service) scaleParams {
	var scale *int
	scale = service.Spec.Replicas

	if service.Status.ComputedReplicas != nil {
		scale = service.Status.ComputedReplicas
	}

	// at one point we told users that -1 meant we don't control scale. nil is now that behavior
	if scale != nil && *scale < 0 {
		scale = nil
	}

	sp := scaleParams{
		MaxSurge:       service.Spec.MaxSurge,
		MaxUnavailable: service.Spec.MaxUnavailable,
	}

	if scale != nil {
		scale32 := int32(*scale)
		sp.Scale = &scale32
	}

	return sp
}
