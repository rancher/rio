package client

const (
	AttachedVolumeType            = "attachedVolume"
	AttachedVolumeFieldDevicePath = "devicePath"
	AttachedVolumeFieldName       = "name"
)

type AttachedVolume struct {
	DevicePath string `json:"devicePath,omitempty" yaml:"devicePath,omitempty"`
	Name       string `json:"name,omitempty" yaml:"name,omitempty"`
}
