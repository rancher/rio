package client

const (
	PodStatusType                       = "podStatus"
	PodStatusFieldConditions            = "conditions"
	PodStatusFieldContainerStatuses     = "containerStatuses"
	PodStatusFieldHostIP                = "hostIP"
	PodStatusFieldInitContainerStatuses = "initContainerStatuses"
	PodStatusFieldMessage               = "message"
	PodStatusFieldNominatedNodeName     = "nominatedNodeName"
	PodStatusFieldPhase                 = "phase"
	PodStatusFieldPodIP                 = "podIP"
	PodStatusFieldQOSClass              = "qosClass"
	PodStatusFieldReason                = "reason"
	PodStatusFieldStartTime             = "startTime"
)

type PodStatus struct {
	Conditions            []PodCondition    `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	ContainerStatuses     []ContainerStatus `json:"containerStatuses,omitempty" yaml:"containerStatuses,omitempty"`
	HostIP                string            `json:"hostIP,omitempty" yaml:"hostIP,omitempty"`
	InitContainerStatuses []ContainerStatus `json:"initContainerStatuses,omitempty" yaml:"initContainerStatuses,omitempty"`
	Message               string            `json:"message,omitempty" yaml:"message,omitempty"`
	NominatedNodeName     string            `json:"nominatedNodeName,omitempty" yaml:"nominatedNodeName,omitempty"`
	Phase                 string            `json:"phase,omitempty" yaml:"phase,omitempty"`
	PodIP                 string            `json:"podIP,omitempty" yaml:"podIP,omitempty"`
	QOSClass              string            `json:"qosClass,omitempty" yaml:"qosClass,omitempty"`
	Reason                string            `json:"reason,omitempty" yaml:"reason,omitempty"`
	StartTime             string            `json:"startTime,omitempty" yaml:"startTime,omitempty"`
}
