package client

const (
	ExternalServiceSpecType         = "externalServiceSpec"
	ExternalServiceSpecFieldSpaceID = "spaceId"
	ExternalServiceSpecFieldStackID = "stackId"
	ExternalServiceSpecFieldTarget  = "target"
)

type ExternalServiceSpec struct {
	SpaceID string `json:"spaceId,omitempty" yaml:"spaceId,omitempty"`
	StackID string `json:"stackId,omitempty" yaml:"stackId,omitempty"`
	Target  string `json:"target,omitempty" yaml:"target,omitempty"`
}
