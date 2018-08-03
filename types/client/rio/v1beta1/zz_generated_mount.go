package client

const (
	MountType               = "mount"
	MountFieldBindOptions   = "bind"
	MountFieldKind          = "kind"
	MountFieldReadOnly      = "readonly"
	MountFieldSource        = "source"
	MountFieldTarget        = "target"
	MountFieldVolumeOptions = "volume"
)

type Mount struct {
	BindOptions   *BindOptions   `json:"bind,omitempty" yaml:"bind,omitempty"`
	Kind          string         `json:"kind,omitempty" yaml:"kind,omitempty"`
	ReadOnly      bool           `json:"readonly,omitempty" yaml:"readonly,omitempty"`
	Source        string         `json:"source,omitempty" yaml:"source,omitempty"`
	Target        string         `json:"target,omitempty" yaml:"target,omitempty"`
	VolumeOptions *VolumeOptions `json:"volume,omitempty" yaml:"volume,omitempty"`
}
