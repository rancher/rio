package schema

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/mapper"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/types/factory"
	rm "github.com/rancher/rio/types/mapper"
)

var (
	Version = types.APIVersion{
		Version:          "v1",
		Group:            "rio.cattle.io",
		Path:             "/v1-rio/project",
		SubContext:       true,
		SubContextSchema: "/v1-rio/schemas/project",
	}

	Schemas = factory.Schemas(&Version).
		Init(configTypes).
		Init(stackTypes).
		Init(serviceTypes).
		Init(volumeTypes).
		Init(routeTypes).
		Init(externalServiceTypes).
		MustImport(&Version, v1.InternalStack{})
)

func volumeTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, v1.VolumeStatus{},
			mapper.Drop{Field: "conditions"},
			&mapper.Embed{Field: "pvcStatus"},
		).
		AddMapperForType(&Version, v1.Volume{},
			mapper.Drop{Field: "namespace"},
			&mapper.Embed{Field: "status"},
			mapper.Drop{Field: "phase"},
		)
}

func configTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, v1.Config{},
			mapper.Drop{Field: "namespace"},
		)
}

func routeTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, v1.RouteSet{},
			mapper.Drop{Field: "namespace"},
		)
}

func serviceTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, v1.ServiceRevision{},
			mapper.Drop{Field: "serviceName"}).
		AddMapperForType(&Version, v1.ServiceSpec{},
			mapper.Move{From: "labels", To: "serviceLabels"},
			&mapper.Embed{Field: "revision"},
			rm.NewMetadata("metadata"),
		).
		AddMapperForType(&Version, v1.ServiceUnversionedSpec{},
			rm.NewMetadata("metadata"),
		).
		AddMapperForType(&Version, v1.ServiceStatus{},
			&rm.DeploymentStatus{},
			mapper.Drop{Field: "deploymentStatus"},
		).
		AddMapperForType(&Version, v1.Service{},
			mapper.PendingStatus{},
			mapper.Drop{Field: "namespace"},
			mapper.Drop{Field: "labels"},
			mapper.Move{From: "serviceLabels", To: "labels"},
		)
}

func stackTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, v1.Stack{},
			mapper.PendingStatus{},
			mapper.Move{From: "namespace", To: "projectId", CodeName: "ProjectID"}).
		MustImportAndCustomize(&Version, v1.Stack{}, func(schema *types.Schema) {
			schema.MustCustomizeField("projectId", func(f types.Field) types.Field {
				f.Type = "reference[project]"
				return f
			})
		})
}

func externalServiceTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.AddMapperForType(&Version, v1.ExternalService{},
		mapper.Drop{Field: "namespace"}).
		MustImport(&Version, v1.ExternalService{})
}
