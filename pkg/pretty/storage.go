package pretty

import (
	"github.com/rancher/norman/types"
	pm "github.com/rancher/rio/pkg/pretty/mapper"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

func storage(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, riov1.VolumeOptions{},
			pm.AliasField{Field: "noCopy", Names: []string{"nocopy"}},
		).
		AddMapperForType(&Version, riov1.Mount{},
			pm.AliasField{Field: "kind", Names: []string{"type"}},
		)
}
