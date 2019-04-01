package v1

type Mount struct {
	Kind string `json:"kind,omitempty" norman:"type=enum,options=bind|volume"`
	// Source specifies the name of the mount. Depending on mount type, this
	// may be a volume name or a host path, or even ignored.
	Source   string `json:"source,omitempty"`
	Target   string `json:"target,omitempty"`
	ReadOnly bool   `json:"readonly,omitempty"`

	BindOptions   *BindOptions   `json:"bind,omitempty"`
	VolumeOptions *VolumeOptions `json:"volume,omitempty"`
}

// Propagation represents the propagation of a mount.
type Propagation string

const (
	// PropagationRPrivate RPRIVATE
	PropagationRPrivate Propagation = "rprivate"
	// PropagationPrivate PRIVATE
	PropagationPrivate Propagation = "private"
	// PropagationRShared RSHARED
	PropagationRShared Propagation = "rshared"
	// PropagationShared SHARED
	PropagationShared Propagation = "shared"
	// PropagationRSlave RSLAVE
	PropagationRSlave Propagation = "rslave"
	// PropagationSlave SLAVE
	PropagationSlave Propagation = "slave"
)

// Propagations is the list of all valid mount propagations
var Propagations = []Propagation{
	PropagationRPrivate,
	PropagationPrivate,
	PropagationRShared,
	PropagationShared,
	PropagationRSlave,
	PropagationSlave,
}

// BindOptions defines options specific to mounts of type "bind".
type BindOptions struct {
	Propagation Propagation `json:"propagation,omitempty"`
}

// VolumeOptions represents the options for a mount of type volume.
type VolumeOptions struct {
	Driver   string `json:"driver,omitempty"`
	SizeInGB int    `json:"sizeInGb,omitempty"`
	SubPath  string `json:"subPath,omitempty"`
}

// Tmpfs defines options specific to mounts of type "tmpfs".
type Tmpfs struct {
	SizeBytes int64  `json:"sizeBytes,omitempty"`
	ReadOnly  bool   `json:"readOnly,omitempty"`
	Path      string `json:"path,omitempty" norman:"required"`
}
