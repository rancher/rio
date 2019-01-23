package client

const (
	ExternalServiceSpecType             = "externalServiceSpec"
	ExternalServiceSpecFieldFQDN        = "fqdn"
	ExternalServiceSpecFieldIPAddresses = "ipAddresses"
	ExternalServiceSpecFieldProjectID   = "projectId"
	ExternalServiceSpecFieldService     = "service"
	ExternalServiceSpecFieldStackID     = "stackId"
)

type ExternalServiceSpec struct {
	FQDN        string   `json:"fqdn,omitempty" yaml:"fqdn,omitempty"`
	IPAddresses []string `json:"ipAddresses,omitempty" yaml:"ipAddresses,omitempty"`
	ProjectID   string   `json:"projectId,omitempty" yaml:"projectId,omitempty"`
	Service     string   `json:"service,omitempty" yaml:"service,omitempty"`
	StackID     string   `json:"stackId,omitempty" yaml:"stackId,omitempty"`
}
