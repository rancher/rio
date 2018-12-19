package client

const (
	ConfigStatusType            = "configStatus"
	ConfigStatusFieldConditions = "conditions"
)

type ConfigStatus struct {
	Conditions []GenericCondition `json:"conditions,omitempty" yaml:"conditions,omitempty"`
}
