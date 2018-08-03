package client

const (
	VolumeSpecType             = "volumeSpec"
	VolumeSpecFieldAccessMode  = "accessMode"
	VolumeSpecFieldDescription = "description"
	VolumeSpecFieldDriver      = "driver"
	VolumeSpecFieldSizeInGB    = "sizeInGb"
	VolumeSpecFieldSpaceID     = "spaceId"
	VolumeSpecFieldStackID     = "stackId"
	VolumeSpecFieldTemplate    = "template"
)

type VolumeSpec struct {
	AccessMode  string `json:"accessMode,omitempty" yaml:"accessMode,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Driver      string `json:"driver,omitempty" yaml:"driver,omitempty"`
	SizeInGB    int64  `json:"sizeInGb,omitempty" yaml:"sizeInGb,omitempty"`
	SpaceID     string `json:"spaceId,omitempty" yaml:"spaceId,omitempty"`
	StackID     string `json:"stackId,omitempty" yaml:"stackId,omitempty"`
	Template    bool   `json:"template,omitempty" yaml:"template,omitempty"`
}
