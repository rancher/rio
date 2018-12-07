package client

const (
	ContainerPortType               = "containerPort"
	ContainerPortFieldContainerPort = "containerPort"
	ContainerPortFieldHostIP        = "hostIP"
	ContainerPortFieldHostPort      = "hostPort"
	ContainerPortFieldName          = "name"
	ContainerPortFieldProtocol      = "protocol"
)

type ContainerPort struct {
	ContainerPort int64  `json:"containerPort,omitempty" yaml:"containerPort,omitempty"`
	HostIP        string `json:"hostIP,omitempty" yaml:"hostIP,omitempty"`
	HostPort      int64  `json:"hostPort,omitempty" yaml:"hostPort,omitempty"`
	Name          string `json:"name,omitempty" yaml:"name,omitempty"`
	Protocol      string `json:"protocol,omitempty" yaml:"protocol,omitempty"`
}
