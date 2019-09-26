package podcontrollers

import (
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type scaleParams struct {
	Scale          int32
	MaxSurge       *intstr.IntOrString
	MaxUnavailable *intstr.IntOrString
}

func parseScaleParams(service *riov1.ServiceSpec) scaleParams {
	scaleNum := 0
	if service.Replicas == nil {
		scaleNum = 1
	}
	if service.Replicas != nil {
		scaleNum = *service.Replicas
	}
	scale := int32(scaleNum)

	return scaleParams{
		Scale:          scale,
		MaxSurge:       service.MaxSurge,
		MaxUnavailable: service.MaxUnavailable,
	}
}
