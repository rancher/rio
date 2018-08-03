package client

const (
	ServiceStatusType             = "serviceStatus"
	ServiceStatusFieldConditions  = "conditions"
	ServiceStatusFieldScaleStatus = "scaleStatus"
)

type ServiceStatus struct {
	Conditions  []Condition  `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	ScaleStatus *ScaleStatus `json:"scaleStatus,omitempty" yaml:"scaleStatus,omitempty"`
}
