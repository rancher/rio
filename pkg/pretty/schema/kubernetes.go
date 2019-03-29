package schema

import (
	"github.com/rancher/mapper"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/pretty/objectmappers"
)

func kubernetes(schemas *mapper.Schemas) *mapper.Schemas {
	return schemas.AddMapperForType(v1.Kubernetes{},
		objectmappers.NewCRDs("customResourceDefinitions"),
		objectmappers.NewCRDs("namespacedCustomResourceDefinitions"),
	)
}
