package client

const (
	ExternalServiceSpecType           = "externalServiceSpec"
	ExternalServiceSpecFieldProjectID = "projectId"
	ExternalServiceSpecFieldStackID   = "stackId"
	ExternalServiceSpecFieldTarget    = "target"
)

type ExternalServiceSpec struct {
	ProjectID string `json:"projectId,omitempty" yaml:"projectId,omitempty"`
	StackID   string `json:"stackId,omitempty" yaml:"stackId,omitempty"`
	Target    string `json:"target,omitempty" yaml:"target,omitempty"`
}
