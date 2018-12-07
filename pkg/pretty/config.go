package pretty

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/mapper"
	mapper2 "github.com/rancher/rio/pkg/pretty/mapper"
	"github.com/rancher/rio/types/client/rio/v1"
)

func config(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, client.Config{},
			mapper.Drop{Field: "projectId"},
			mapper.Drop{Field: "stackId"},
			mapper2.ConfigContent{},
		).
		MustImport(&Version, client.Config{}, struct {
			File string `json:"file,omitempty"`
		}{})
}
