package client

const (
	ExposedPortType            = "exposedPort"
	ExposedPortFieldIP         = "ip"
	ExposedPortFieldName       = "name"
	ExposedPortFieldPort       = "port"
	ExposedPortFieldProtocol   = "protocol"
	ExposedPortFieldTargetPort = "targetPort"
)

type ExposedPort struct {
	IP         string `json:"ip,omitempty" yaml:"ip,omitempty"`
	Name       string `json:"name,omitempty" yaml:"name,omitempty"`
	Port       int64  `json:"port,omitempty" yaml:"port,omitempty"`
	Protocol   string `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	TargetPort int64  `json:"targetPort,omitempty" yaml:"targetPort,omitempty"`
}
