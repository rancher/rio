package podcontrollers

import (
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type scaleParams struct {
	Scale          int32
	MaxSurge       *intstr.IntOrString
	MaxUnavailable *intstr.IntOrString
	BatchSize      int
}

func parseScaleParams(service *riov1.ServiceSpec) scaleParams {
	scale := int32(service.Scale)
	batchSize := service.UpdateBatchSize

	if scale == 0 {
		scale = 1
	}

	if batchSize == 0 {
		batchSize = 1
	}

	if int32(batchSize) > scale {
		batchSize = int(scale)
	}

	maxSurge := intstr.FromInt(batchSize)
	maxUnavailable := intstr.FromInt(0)

	return scaleParams{
		Scale:          scale,
		MaxSurge:       &maxSurge,
		MaxUnavailable: &maxUnavailable,
		BatchSize:      batchSize,
	}
}
