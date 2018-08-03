package client

const (
	VolumeOptionsType          = "volumeOptions"
	VolumeOptionsFieldDriver   = "driver"
	VolumeOptionsFieldNoCopy   = "noCopy"
	VolumeOptionsFieldSizeInGB = "sizeInGb"
	VolumeOptionsFieldSubPath  = "subPath"
)

type VolumeOptions struct {
	Driver   string `json:"driver,omitempty" yaml:"driver,omitempty"`
	NoCopy   bool   `json:"noCopy,omitempty" yaml:"noCopy,omitempty"`
	SizeInGB int64  `json:"sizeInGb,omitempty" yaml:"sizeInGb,omitempty"`
	SubPath  string `json:"subPath,omitempty" yaml:"subPath,omitempty"`
}
