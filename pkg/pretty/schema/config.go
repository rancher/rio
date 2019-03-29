package schema

import (
	"github.com/rancher/mapper"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	mapper2 "github.com/rancher/rio/pkg/pretty/mapper"
)

func config(schemas *mapper.Schemas) *mapper.Schemas {
	return schemas.
		AddMapperForType(riov1.ConfigSpec{},
			mapper2.ConfigContent{},
		).
		MustImport(riov1.ConfigSpec{}, struct {
			File string `json:"file,omitempty"`
		}{})
}
