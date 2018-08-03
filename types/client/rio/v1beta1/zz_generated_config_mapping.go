package client

const (
	ConfigMappingType        = "configMapping"
	ConfigMappingFieldGID    = "gid"
	ConfigMappingFieldMode   = "mode"
	ConfigMappingFieldSource = "source"
	ConfigMappingFieldTarget = "target"
	ConfigMappingFieldUID    = "uid"
)

type ConfigMapping struct {
	GID    int64  `json:"gid,omitempty" yaml:"gid,omitempty"`
	Mode   string `json:"mode,omitempty" yaml:"mode,omitempty"`
	Source string `json:"source,omitempty" yaml:"source,omitempty"`
	Target string `json:"target,omitempty" yaml:"target,omitempty"`
	UID    int64  `json:"uid,omitempty" yaml:"uid,omitempty"`
}
