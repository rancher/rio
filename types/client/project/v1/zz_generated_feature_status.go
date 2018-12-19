package client

const (
	FeatureStatusType            = "featureStatus"
	FeatureStatusFieldConditions = "conditions"
)

type FeatureStatus struct {
	Conditions []GenericCondition `json:"conditions,omitempty" yaml:"conditions,omitempty"`
}
