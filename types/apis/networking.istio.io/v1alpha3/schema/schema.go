package schema

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3"
	"github.com/rancher/rio/types/factory"
)

var (
	Version = types.APIVersion{
		Version: "v1alpha3",
		Group:   "networking.istio.io",
		Path:    "/v1alpha3-istio-networking",
	}

	Schemas = factory.Schemas(&Version).
		MustImport(&Version, v1alpha3.Gateway{}).
		MustImport(&Version, v1alpha3.VirtualService{})
)
