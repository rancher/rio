package pretty

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/mapper"
	mapper2 "github.com/rancher/rio/pkg/pretty/mapper"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

func volume(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, riov1.VolumeSpec{},
			mapper.Move{From: "sizeInGb", To: "size"},
			mapper2.AliasField{Field: "size", Names: []string{"sizeInGb"}},
		)
}
