package pretty

import (
	"github.com/rancher/norman/types"
	pm "github.com/rancher/rio/pkg/pretty/mapper"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

func storage(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, client.VolumeOptions{},
			pm.AliasField{Field: "noCopy", Names: []string{"nocopy"}},
		).
		AddMapperForType(&Version, client.Mount{},
			pm.AliasField{Field: "kind", Names: []string{"type"}},
		)
}
