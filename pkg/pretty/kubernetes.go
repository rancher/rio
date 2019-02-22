package pretty

import (
	"github.com/rancher/norman/types"
	pm "github.com/rancher/rio/pkg/pretty/mapper"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

func kubernetes(schemas *types.Schemas) *types.Schemas {
	return schemas.AddMapperForType(&Version, riov1.Kubernetes{},
		pm.NewCRDs("customResourceDefinitions"),
		pm.NewCRDs("namespacedCustomResourceDefinitions"),
	)
}
