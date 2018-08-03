package client

const (
	CustomResourceDefinitionType         = "customResourceDefinition"
	CustomResourceDefinitionFieldGroup   = "group"
	CustomResourceDefinitionFieldKind    = "kind"
	CustomResourceDefinitionFieldVersion = "version"
)

type CustomResourceDefinition struct {
	Group   string `json:"group,omitempty" yaml:"group,omitempty"`
	Kind    string `json:"kind,omitempty" yaml:"kind,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
}
