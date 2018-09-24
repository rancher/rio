package client

const (
	ServiceStatusType                   = "serviceStatus"
	ServiceStatusFieldConditions        = "conditions"
	ServiceStatusFieldDaemonSetStatus   = "daemonSetStatus"
	ServiceStatusFieldEndpoints         = "endpoint"
	ServiceStatusFieldScaleStatus       = "scaleStatus"
	ServiceStatusFieldStatefulSetStatus = "statefulSetStatus"
)

type ServiceStatus struct {
	Conditions        []Condition        `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	DaemonSetStatus   *DaemonSetStatus   `json:"daemonSetStatus,omitempty" yaml:"daemonSetStatus,omitempty"`
	Endpoints         []Endpoint         `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	ScaleStatus       *ScaleStatus       `json:"scaleStatus,omitempty" yaml:"scaleStatus,omitempty"`
	StatefulSetStatus *StatefulSetStatus `json:"statefulSetStatus,omitempty" yaml:"statefulSetStatus,omitempty"`
}
