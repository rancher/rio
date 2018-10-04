package schema

import (
	"github.com/rancher/norman/types"
	m "github.com/rancher/norman/types/mapper"
	"github.com/rancher/types/factory"
	"k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
)

var (
	Version = types.APIVersion{
		Version:          "v3",
		Group:            "cluster.cattle.io",
		Path:             "/v3/cluster",
		SubContext:       true,
		SubContextSchema: "/v3/schemas/cluster",
	}

	Schemas = factory.Schemas(&Version).
		Init(namespaceTypes).
		Init(persistentVolumeTypes).
		Init(storageClassTypes)
)

func namespaceTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, v1.NamespaceSpec{},
			&m.Drop{Field: "finalizers"},
		).
		AddMapperForType(&Version, v1.Namespace{},
			&m.AnnotationField{Field: "description"},
			&m.AnnotationField{Field: "projectId"},
			&m.AnnotationField{Field: "resourceQuota", Object: true},
			&m.Drop{Field: "status"},
		).
		MustImport(&Version, NamespaceResourceQuota{}).
		MustImport(&Version, v1.Namespace{}, struct {
			Description   string `json:"description"`
			ProjectID     string `norman:"type=reference[/v3/schemas/project],noupdate"`
			ResourceQuota string `json:"resourceQuota,omitempty" norman:"type=namespaceResourceQuota"`
		}{}).
		MustImport(&Version, NamespaceMove{}).
		MustImportAndCustomize(&Version, v1.Namespace{}, func(schema *types.Schema) {
			schema.ResourceActions["move"] = types.Action{
				Input: "namespaceMove",
			}
		})
}

func persistentVolumeTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, v1.PersistentVolume{},
			&m.AnnotationField{Field: "description"},
		).
		AddMapperForType(&Version, v1.HostPathVolumeSource{},
			m.Move{From: "type", To: "kind"},
			m.Enum{
				Options: []string{
					"DirectoryOrCreate",
					"Directory",
					"FileOrCreate",
					"File",
					"Socket",
					"CharDevice",
					"BlockDevice",
				},
				Field: "kind",
			},
		).
		MustImport(&Version, v1.PersistentVolumeSpec{}, struct {
			StorageClassName *string `json:"storageClassName,omitempty" norman:"type=reference[storageClass]"`
		}{}).
		MustImport(&Version, v1.PersistentVolume{}, struct {
			Description string `json:"description"`
		}{})
}

func storageClassTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, storagev1.StorageClass{},
			&m.AnnotationField{Field: "description"},
		).
		MustImport(&Version, storagev1.StorageClass{}, struct {
			Description   string `json:"description"`
			ReclaimPolicy string `json:"reclaimPolicy,omitempty" norman:"type=enum,options=Recycle|Delete|Retain"`
		}{})
}
