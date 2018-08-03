package schema

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/mapper"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"github.com/rancher/rio/types/factory"
	rm "github.com/rancher/rio/types/mapper"
	mapper2 "github.com/rancher/types/mapper"
)

var (
	Version = types.APIVersion{
		Version:          "v1beta1",
		Group:            "rio.cattle.io",
		Path:             "/v1beta1-rio/spaces",
		SubContext:       true,
		SubContextSchema: "/v1beta1-rio/schemas/space",
	}

	Schemas = factory.Schemas(&Version).
		Init(configTypes).
		Init(stackTypes).
		Init(serviceTypes).
		Init(volumeTypes).
		Init(routeTypes).
		MustImport(&Version, v1beta1.InternalStack{})
)

func volumeTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, v1beta1.VolumeStatus{},
			mapper.Drop{Field: "conditions"},
			&mapper.Embed{Field: "pvcStatus"},
		).
		AddMapperForType(&Version, v1beta1.Volume{},
			mapper.Drop{Field: "namespace"},
			&mapper.Embed{Field: "status"},
			mapper.Drop{Field: "phase"},
		)
}

func configTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, v1beta1.Config{},
			mapper.Drop{Field: "namespace"},
		)
}

func routeTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, v1beta1.RouteSet{},
			mapper.Drop{Field: "namespace"},
		)
}

func serviceTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, v1beta1.ServiceRevision{},
			mapper2.Status{},
			&mapper.Embed{Field: "spec"},
			&mapper.ReadOnly{
				Field:     "status",
				SubFields: true,
			},
			&mapper.Embed{Field: "status"},
		).
		AddMapperForType(&Version, v1beta1.ServiceSpec{},
			mapper.Move{From: "labels", To: "serviceLabels"},
			rm.NewMetadata("metadata"),
		).
		AddMapperForType(&Version, v1beta1.ServiceUnversionedSpec{},
			rm.NewMetadata("metadata"),
		).
		AddMapperForType(&Version, v1beta1.ServiceStatus{},
			&rm.DeploymentStatus{},
			mapper.Drop{Field: "deploymentStatus"},
		).
		AddMapperForType(&Version, v1beta1.Service{},
			mapper.Drop{Field: "namespace"},
			mapper.Drop{Field: "labels"},
			mapper.Move{From: "serviceLabels", To: "labels"},
		)
}

func stackTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, v1beta1.Stack{},
			mapper.Move{From: "namespace", To: "spaceId", CodeName: "SpaceID"}).
		MustImportAndCustomize(&Version, v1beta1.Stack{}, func(schema *types.Schema) {
			schema.MustCustomizeField("spaceId", func(f types.Field) types.Field {
				f.Type = "reference[space]"
				return f
			})
		})
}
