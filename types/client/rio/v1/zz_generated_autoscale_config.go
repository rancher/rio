package client

const (
	AutoscaleConfigType             = "autoscaleConfig"
	AutoscaleConfigFieldConcurrency = "concurrency"
	AutoscaleConfigFieldMaxScale    = "maxScale"
	AutoscaleConfigFieldMinScale    = "minScale"
)

type AutoscaleConfig struct {
	Concurrency int64 `json:"concurrency,omitempty" yaml:"concurrency,omitempty"`
	MaxScale    int64 `json:"maxScale,omitempty" yaml:"maxScale,omitempty"`
	MinScale    int64 `json:"minScale,omitempty" yaml:"minScale,omitempty"`
}
