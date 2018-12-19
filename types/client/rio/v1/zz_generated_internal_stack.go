package client

const (
	InternalStackType                  = "internalStack"
	InternalStackFieldConfigs          = "configs"
	InternalStackFieldExternalServices = "externalservices"
	InternalStackFieldKubernetes       = "kubernetes"
	InternalStackFieldRoutes           = "routes"
	InternalStackFieldServices         = "services"
	InternalStackFieldVolumes          = "volumes"
)

type InternalStack struct {
	Configs          map[string]Config          `json:"configs,omitempty" yaml:"configs,omitempty"`
	ExternalServices map[string]ExternalService `json:"externalservices,omitempty" yaml:"externalservices,omitempty"`
	Kubernetes       *Kubernetes                `json:"kubernetes,omitempty" yaml:"kubernetes,omitempty"`
	Routes           map[string]RouteSet        `json:"routes,omitempty" yaml:"routes,omitempty"`
	Services         map[string]Service         `json:"services,omitempty" yaml:"services,omitempty"`
	Volumes          map[string]Volume          `json:"volumes,omitempty" yaml:"volumes,omitempty"`
}
