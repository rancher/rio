package schema

import (
	"github.com/rancher/mapper"
	"github.com/rancher/mapper/mappers"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	pm "github.com/rancher/rio/pkg/pretty/mapper"
)

var (
	Schemas = newSchemas().
		Init(health).
		Init(storage).
		Init(services).
		Init(config).
		Init(volume).
		Init(route).
		Init(kubernetes).
		AddMapperForType(Stack{},
			pm.RouteSet{Field: "routes"},
			pm.RevisionMapper{Field: "services"},
		).
		MustImport(Stack{}).
		MustImport(v1.Service{})
)

func newSchemas() *mapper.Schemas {
	schemas := mapper.NewSchemas()
	schemas.DefaultPostMappers = func() []mapper.Mapper {
		return []mapper.Mapper{
			mappers.JSONKeys{},
			mappers.Drop{
				Field:            "type",
				IgnoreDefinition: true,
			},
		}
	}
	return schemas
}

type TemplateMeta struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
	IconURL string `json:"iconUrl,omitempty"`
	Readme  string `json:"readme,omitempty"`
}

type Stack struct {
	Meta       TemplateMeta              `json:"meta,omitempty"`
	Services   map[string]v1.ServiceSpec `json:"services,omitempty"`
	Configs    map[string]v1.ConfigSpec  `json:"configs,omitempty"`
	Volumes    map[string]v1.VolumeSpec  `json:"volumes,omitempty"`
	Routes     map[string]v1.RouterSpec  `json:"routes,omitempty"`
	Questions  []v1.Question             `json:"questions,omitempty"`
	Kubernetes v1.Kubernetes             `json:"kubernetes,omitempty"`
}
