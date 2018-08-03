package client

const (
	NodeSpecType               = "nodeSpec"
	NodeSpecFieldConfigSource  = "configSource"
	NodeSpecFieldExternalID    = "externalID"
	NodeSpecFieldPodCIDR       = "podCIDR"
	NodeSpecFieldProviderID    = "providerID"
	NodeSpecFieldTaints        = "taints"
	NodeSpecFieldUnschedulable = "unschedulable"
)

type NodeSpec struct {
	ConfigSource  *NodeConfigSource `json:"configSource,omitempty" yaml:"configSource,omitempty"`
	ExternalID    string            `json:"externalID,omitempty" yaml:"externalID,omitempty"`
	PodCIDR       string            `json:"podCIDR,omitempty" yaml:"podCIDR,omitempty"`
	ProviderID    string            `json:"providerID,omitempty" yaml:"providerID,omitempty"`
	Taints        []Taint           `json:"taints,omitempty" yaml:"taints,omitempty"`
	Unschedulable bool              `json:"unschedulable,omitempty" yaml:"unschedulable,omitempty"`
}
