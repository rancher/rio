package schema

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/rio/types/apis/build.knative.dev/v1alpha1"
	"github.com/rancher/rio/types/factory"
)

var (
	Version = types.APIVersion{
		Version: "v1alpha1",
		Group:   "build.knative.dev",
		Path:    "/v1alpha1-build-knative",
	}

	Schemas = factory.Schemas(&Version).
		MustImport(&Version, v1alpha1.Build{})
)
