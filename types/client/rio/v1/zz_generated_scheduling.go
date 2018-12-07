package client

const (
	SchedulingType           = "scheduling"
	SchedulingFieldNode      = "node"
	SchedulingFieldScheduler = "scheduler"
)

type Scheduling struct {
	Node      *NodeScheduling `json:"node,omitempty" yaml:"node,omitempty"`
	Scheduler string          `json:"scheduler,omitempty" yaml:"scheduler,omitempty"`
}
