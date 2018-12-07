package client

const (
	FaultType             = "fault"
	FaultFieldAbort       = "abort"
	FaultFieldDelayMillis = "delayMillis"
	FaultFieldPercentage  = "percentage"
)

type Fault struct {
	Abort       *Abort `json:"abort,omitempty" yaml:"abort,omitempty"`
	DelayMillis int64  `json:"delayMillis,omitempty" yaml:"delayMillis,omitempty"`
	Percentage  int64  `json:"percentage,omitempty" yaml:"percentage,omitempty"`
}
