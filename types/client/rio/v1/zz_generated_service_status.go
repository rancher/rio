package client

const (
	ServiceStatusType                   = "serviceStatus"
	ServiceStatusFieldConditions        = "conditions"
	ServiceStatusFieldDaemonSetStatus   = "daemonSetStatus"
	ServiceStatusFieldEndpoints         = "endpoints"
	ServiceStatusFieldScaleStatus       = "scaleStatus"
	ServiceStatusFieldStatefulSetStatus = "statefulSetStatus"
)

type ServiceStatus struct {
	Conditions        []Condition        `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	DaemonSetStatus   *DaemonSetStatus   `json:"daemonSetStatus,omitempty" yaml:"daemonSetStatus,omitempty"`
	Endpoints         []Endpoint         `json:"endpoints,omitempty" yaml:"endpoints,omitempty"`
	ScaleStatus       *ScaleStatus       `json:"scaleStatus,omitempty" yaml:"scaleStatus,omitempty"`
	StatefulSetStatus *StatefulSetStatus `json:"statefulSetStatus,omitempty" yaml:"statefulSetStatus,omitempty"`
}
