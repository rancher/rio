package schema

import (
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "k8s.io/api/core/v1"
)

type Kubernetes struct {
	Manifest string `json:"manifest,omitempty"`
}

type ExternalRiofile struct {
	Services         map[string]riov1.Service         `json:"services,omitempty"`
	Configs          map[string]v1.ConfigMap          `json:"configs,omitempty"`
	Routers          map[string]riov1.Router          `json:"routers,omitempty"`
	ExternalServices map[string]riov1.ExternalService `json:"externalservices,omitempty"`
	Kubernetes       *Kubernetes                      `json:"kubernetes,omitempty"`
}
