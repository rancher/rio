package deploy

import (
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type scaleParams struct {
	Scale          int32
	MaxSurge       *intstr.IntOrString
	MaxUnavailable *intstr.IntOrString
	BatchSize      int
}

func parseScaleParams(service *v1beta1.ServiceUnversionedSpec) scaleParams {
	scale := int32(service.Scale)
	batchSize := service.BatchSize

	if batchSize == 0 {
		batchSize = 1
	}

	if int32(batchSize) > scale {
		batchSize = int(scale)
	}

	surge := batchSize
	unavailable := 0

	if service.UpdateOrder == "start-first" {
		surge = batchSize
		unavailable = 0
	} else if service.UpdateOrder == "stop-first" {
		surge = 0
		unavailable = batchSize
	}

	maxSurge := intstr.FromInt(surge)
	maxUnavailable := intstr.FromInt(unavailable)

	return scaleParams{
		Scale:          scale,
		MaxSurge:       &maxSurge,
		MaxUnavailable: &maxUnavailable,
		BatchSize:      batchSize,
	}
}
