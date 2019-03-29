package schema

import (
	"github.com/rancher/mapper"
	m "github.com/rancher/mapper/mappers"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

func storage(schemas *mapper.Schemas) *mapper.Schemas {
	return schemas.
		AddMapperForType(riov1.VolumeOptions{},
			m.AliasField{Field: "noCopy", Names: []string{"nocopy"}},
		).
		AddMapperForType(riov1.Mount{},
			m.AliasField{Field: "kind", Names: []string{"type"}},
		)
}
