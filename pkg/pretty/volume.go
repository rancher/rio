package pretty

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/mapper"
	mapper2 "github.com/rancher/rio/pkg/pretty/mapper"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

func volume(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, client.Volume{},
			mapper.Drop{Field: "spaceId"},
			mapper.Drop{Field: "stackId"},
			mapper.Move{From: "sizeInGb", To: "size"},
			mapper2.AliasField{Field: "size", Names: []string{"sizeInGb"}},
		)
}
