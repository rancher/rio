package client

const (
	VolumeSpecType             = "volumeSpec"
	VolumeSpecFieldAccessMode  = "accessMode"
	VolumeSpecFieldDescription = "description"
	VolumeSpecFieldDriver      = "driver"
	VolumeSpecFieldProjectID   = "projectId"
	VolumeSpecFieldSizeInGB    = "sizeInGb"
	VolumeSpecFieldStackID     = "stackId"
	VolumeSpecFieldTemplate    = "template"
)

type VolumeSpec struct {
	AccessMode  string `json:"accessMode,omitempty" yaml:"accessMode,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Driver      string `json:"driver,omitempty" yaml:"driver,omitempty"`
	ProjectID   string `json:"projectId,omitempty" yaml:"projectId,omitempty"`
	SizeInGB    int64  `json:"sizeInGb,omitempty" yaml:"sizeInGb,omitempty"`
	StackID     string `json:"stackId,omitempty" yaml:"stackId,omitempty"`
	Template    bool   `json:"template,omitempty" yaml:"template,omitempty"`
}
