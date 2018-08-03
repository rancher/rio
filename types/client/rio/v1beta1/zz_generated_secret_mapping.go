package client

const (
	SecretMappingType        = "secretMapping"
	SecretMappingFieldMode   = "mode"
	SecretMappingFieldSource = "source"
	SecretMappingFieldTarget = "target"
)

type SecretMapping struct {
	Mode   string `json:"mode,omitempty" yaml:"mode,omitempty"`
	Source string `json:"source,omitempty" yaml:"source,omitempty"`
	Target string `json:"target,omitempty" yaml:"target,omitempty"`
}
