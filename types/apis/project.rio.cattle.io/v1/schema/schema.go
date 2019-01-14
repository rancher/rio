package schema

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/mapper"
	projectv1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	"github.com/rancher/rio/types/factory"
	typemapper "github.com/rancher/types/mapper"
	v1 "k8s.io/api/core/v1"
)

var (
	Version = types.APIVersion{
		Version: "v1",
		Group:   "project.rio.cattle.io",
		Path:    "/v1-rio",
	}

	Schemas = factory.Schemas(&Version).
		MustImport(&Version, projectv1.ListenConfig{}).
		MustImport(&Version, v1.Node{}).
		MustImport(&Version, projectv1.Setting{}).
		Init(podTypes).
		Init(projectTypes).
		Init(publicDomainTypes).
		Init(featureTypes)
)

func podTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, v1.PodTemplateSpec{},
			&mapper.Embed{Field: "spec"},
		).
		AddMapperForType(&Version, v1.Pod{},
			typemapper.ContainerStatus{},
		).
		MustImport(&Version, v1.Container{}, struct {
			State                string
			Transitioning        string
			TransitioningMessage string
			RestartCount         int
			ExitCode             *int
		}{}).
		MustImport(&Version, v1.Pod{}, struct {
			types.Namespaced
		}{})
}

func projectTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.
		TypeName("project", v1.Namespace{}).
		AddMapperForType(&Version, v1.NamespaceSpec{},
			mapper.Drop{Field: "finalizers"},
		).
		AddMapperForType(&Version, v1.NamespaceStatus{},
			mapper.Drop{Field: "phase"},
		).
		AddMapperForType(&Version, v1.Namespace{},
			mapper.LabelField{Field: "displayName"},
			mapper.DisplayName{},
			mapper.Access{Fields: map[string]string{
				"id":   "r",
				"name": "cr",
			}},
		).
		MustImportAndCustomize(&Version, v1.Namespace{}, func(schema *types.Schema) {
			schema.CodeName = "Project"
			schema.CodeNamePlural = "Projects"
		}, struct {
			DisplayName string
		}{},
		)
}

func publicDomainTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.MustImport(&Version, projectv1.PublicDomain{})
}

func featureTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.MustImport(&Version, projectv1.Feature{})
}
