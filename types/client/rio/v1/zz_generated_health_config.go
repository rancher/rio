package client

const (
	HealthConfigType                     = "healthConfig"
	HealthConfigFieldHealthyThreshold    = "healthyThreshold"
	HealthConfigFieldInitialDelaySeconds = "initialDelaySeconds"
	HealthConfigFieldIntervalSeconds     = "intervalSeconds"
	HealthConfigFieldTest                = "test"
	HealthConfigFieldTimeoutSeconds      = "timeoutSeconds"
	HealthConfigFieldUnhealthyThreshold  = "unhealthyThreshold"
)

type HealthConfig struct {
	HealthyThreshold    int64    `json:"healthyThreshold,omitempty" yaml:"healthyThreshold,omitempty"`
	InitialDelaySeconds int64    `json:"initialDelaySeconds,omitempty" yaml:"initialDelaySeconds,omitempty"`
	IntervalSeconds     int64    `json:"intervalSeconds,omitempty" yaml:"intervalSeconds,omitempty"`
	Test                []string `json:"test,omitempty" yaml:"test,omitempty"`
	TimeoutSeconds      int64    `json:"timeoutSeconds,omitempty" yaml:"timeoutSeconds,omitempty"`
	UnhealthyThreshold  int64    `json:"unhealthyThreshold,omitempty" yaml:"unhealthyThreshold,omitempty"`
}
