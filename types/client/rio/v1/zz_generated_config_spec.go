package client

const (
	ConfigSpecType             = "configSpec"
	ConfigSpecFieldContent     = "content"
	ConfigSpecFieldDescription = "description"
	ConfigSpecFieldEncoded     = "encoded"
	ConfigSpecFieldProjectID   = "projectId"
	ConfigSpecFieldStackID     = "stackId"
)

type ConfigSpec struct {
	Content     string `json:"content,omitempty" yaml:"content,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Encoded     bool   `json:"encoded,omitempty" yaml:"encoded,omitempty"`
	ProjectID   string `json:"projectId,omitempty" yaml:"projectId,omitempty"`
	StackID     string `json:"stackId,omitempty" yaml:"stackId,omitempty"`
}
