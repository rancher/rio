package schema

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/mapper"
	"github.com/rancher/rio/types/apis/space.cattle.io/v1beta1"
	"github.com/rancher/rio/types/factory"
	typemapper "github.com/rancher/types/mapper"
	"k8s.io/api/core/v1"
)

var (
	Version = types.APIVersion{
		Version: "v1beta1",
		Group:   "space.cattle.io",
		Path:    "/v1beta1-rio",
	}

	Schemas = factory.Schemas(&Version).
		MustImport(&Version, v1beta1.ListenConfig{}).
		MustImport(&Version, v1.Node{}).
		Init(podTypes).
		Init(spaceTypes)
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

func spaceTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.
		TypeName("space", v1.Namespace{}).
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
			schema.CodeName = "Space"
			schema.CodeNamePlural = "Spaces"
		}, struct {
			DisplayName string
		}{},
		)
}
