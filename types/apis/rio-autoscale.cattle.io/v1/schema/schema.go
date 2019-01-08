package schema

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/rio/types/apis/rio-autoscale.cattle.io/v1"
	"github.com/rancher/rio/types/factory"
)

var (
	APIVersion = types.APIVersion{
		Group:   "rio-autoscale.cattle.io",
		Version: "v1",
		Path:    "/v1-rio-autoscale",
	}
	Schemas = factory.
		Schemas(&APIVersion).
		MustImport(&APIVersion, v1.ServiceScaleRecommendation{})
)
