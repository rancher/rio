package pretty

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/mapper"
	pm "github.com/rancher/rio/pkg/pretty/mapper"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	"github.com/rancher/types/apis/management.cattle.io/v3"
)

var (
	Version = types.APIVersion{
		Version: "v1beta1",
		Group:   "export.cattle.io",
		Path:    "/v1beta1-export",
	}

	Schemas = newSchemas().
		Init(health).
		Init(storage).
		Init(services).
		Init(config).
		Init(volume).
		Init(route).
		Init(kubernetes).
		AddMapperForType(&Version, Stack{},
			pm.RouteSet{Field: "routes"},
		).
		MustImport(&Version, Stack{})
)

func newSchemas() *types.Schemas {
	schemas := types.NewSchemas()
	schemas.DefaultPostMappers = func() []types.Mapper {
		return []types.Mapper{
			pm.JSONKeys{},
			mapper.Drop{
				Field:            "type",
				IgnoreDefinition: true,
			},
		}
	}
	return schemas
}

type TemplateMeta struct {
	Name       string `json:"name,omitempty"`
	Version    string `json:"version,omitempty"`
	IconURL    string `json:"iconUrl,omitempty"`
	RioVersion string `json:"rioVersion,omitempty"`
	Readme     string `json:"readme,omitempty"`
}

type Stack struct {
	Meta       TemplateMeta               `json:"meta,omitempty"`
	Services   map[string]client.Service  `json:"services,omitempty"`
	Configs    map[string]client.Config   `json:"configs,omitempty"`
	Volumes    map[string]client.Volume   `json:"volumes,omitempty"`
	Routes     map[string]client.RouteSet `json:"routes,omitempty"`
	Questions  []v3.Question              `json:"questions,omitempty"`
	Kubernetes client.Kubernetes          `json:"kubernetes,omitempty"`
}
