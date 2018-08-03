package v1beta1

import (
	"fmt"
)

type Kubernetes struct {
	CustomResourceDefinitions           []CustomResourceDefinition `json:"customResourceDefinitions,omitempty"`
	NamespacedCustomResourceDefinitions []CustomResourceDefinition `json:"namespacedCustomResourceDefinitions,omitempty"`
	Manifest                            string                     `json:"manifest,omitempty"`
	NamespacedManifest                  string                     `json:"namespacedManifest,omitempty"`
}

type CustomResourceDefinition struct {
	Kind    string `json:"kind,omitempty"`
	Group   string `json:"group,omitempty"`
	Version string `json:"version,omitempty"`
}

func (n *CustomResourceDefinition) MaybeString() interface{} {
	return fmt.Sprintf("%s.%s/%s", n.Kind, n.Group, n.Version)
}
