package client

const (
	NodeStatusType                 = "nodeStatus"
	NodeStatusFieldAddresses       = "addresses"
	NodeStatusFieldAllocatable     = "allocatable"
	NodeStatusFieldCapacity        = "capacity"
	NodeStatusFieldConditions      = "conditions"
	NodeStatusFieldDaemonEndpoints = "daemonEndpoints"
	NodeStatusFieldImages          = "images"
	NodeStatusFieldNodeInfo        = "nodeInfo"
	NodeStatusFieldPhase           = "phase"
	NodeStatusFieldVolumesAttached = "volumesAttached"
	NodeStatusFieldVolumesInUse    = "volumesInUse"
)

type NodeStatus struct {
	Addresses       []NodeAddress        `json:"addresses,omitempty" yaml:"addresses,omitempty"`
	Allocatable     map[string]string    `json:"allocatable,omitempty" yaml:"allocatable,omitempty"`
	Capacity        map[string]string    `json:"capacity,omitempty" yaml:"capacity,omitempty"`
	Conditions      []NodeCondition      `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	DaemonEndpoints *NodeDaemonEndpoints `json:"daemonEndpoints,omitempty" yaml:"daemonEndpoints,omitempty"`
	Images          []ContainerImage     `json:"images,omitempty" yaml:"images,omitempty"`
	NodeInfo        *NodeSystemInfo      `json:"nodeInfo,omitempty" yaml:"nodeInfo,omitempty"`
	Phase           string               `json:"phase,omitempty" yaml:"phase,omitempty"`
	VolumesAttached []AttachedVolume     `json:"volumesAttached,omitempty" yaml:"volumesAttached,omitempty"`
	VolumesInUse    []string             `json:"volumesInUse,omitempty" yaml:"volumesInUse,omitempty"`
}
