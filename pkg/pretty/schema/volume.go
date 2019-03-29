package schema

import (
	"github.com/rancher/mapper"
	m "github.com/rancher/mapper/mappers"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

func volume(schemas *mapper.Schemas) *mapper.Schemas {
	return schemas.
		AddMapperForType(riov1.VolumeSpec{},
			m.Move{From: "sizeInGb", To: "size"},
			m.AliasField{Field: "size", Names: []string{"sizeInGb"}},
		)
}
