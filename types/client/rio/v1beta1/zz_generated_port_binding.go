package client

const (
	PortBindingType            = "portBinding"
	PortBindingFieldIP         = "ip"
	PortBindingFieldPort       = "port"
	PortBindingFieldProtocol   = "protocol"
	PortBindingFieldTargetPort = "targetPort"
)

type PortBinding struct {
	IP         string `json:"ip,omitempty" yaml:"ip,omitempty"`
	Port       int64  `json:"port,omitempty" yaml:"port,omitempty"`
	Protocol   string `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	TargetPort int64  `json:"targetPort,omitempty" yaml:"targetPort,omitempty"`
}
