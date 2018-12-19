package client

const (
	ExternalServiceStatusType            = "externalServiceStatus"
	ExternalServiceStatusFieldConditions = "conditions"
)

type ExternalServiceStatus struct {
	Conditions []GenericCondition `json:"conditions,omitempty" yaml:"conditions,omitempty"`
}
