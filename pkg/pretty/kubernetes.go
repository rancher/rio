package pretty

import (
	"github.com/rancher/norman/types"
	pm "github.com/rancher/rio/pkg/pretty/mapper"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

func kubernetes(schemas *types.Schemas) *types.Schemas {
	return schemas.AddMapperForType(&Version, client.Kubernetes{},
		pm.NewCRDs("customResourceDefinitions"),
		pm.NewCRDs("namespacedCustomResourceDefinitions"),
	)
}
