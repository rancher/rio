package client

const (
	BindOptionsType             = "bindOptions"
	BindOptionsFieldPropagation = "propagation"
)

type BindOptions struct {
	Propagation string `json:"propagation,omitempty" yaml:"propagation,omitempty"`
}
