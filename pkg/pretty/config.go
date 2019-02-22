package pretty

import (
	"github.com/rancher/norman/types"
	mapper2 "github.com/rancher/rio/pkg/pretty/mapper"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

func config(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, riov1.ConfigSpec{},
			mapper2.ConfigContent{},
		).
		MustImport(&Version, riov1.ConfigSpec{}, struct {
			File string `json:"file,omitempty"`
		}{})
}
