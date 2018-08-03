package client

const (
	ConfigSpecType             = "configSpec"
	ConfigSpecFieldContent     = "content"
	ConfigSpecFieldDescription = "description"
	ConfigSpecFieldEncoded     = "encoded"
	ConfigSpecFieldSpaceID     = "spaceId"
	ConfigSpecFieldStackID     = "stackId"
)

type ConfigSpec struct {
	Content     string `json:"content,omitempty" yaml:"content,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Encoded     bool   `json:"encoded,omitempty" yaml:"encoded,omitempty"`
	SpaceID     string `json:"spaceId,omitempty" yaml:"spaceId,omitempty"`
	StackID     string `json:"stackId,omitempty" yaml:"stackId,omitempty"`
}
