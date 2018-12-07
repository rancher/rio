package client

const (
	TmpfsType           = "tmpfs"
	TmpfsFieldPath      = "path"
	TmpfsFieldReadOnly  = "readOnly"
	TmpfsFieldSizeBytes = "sizeBytes"
)

type Tmpfs struct {
	Path      string `json:"path,omitempty" yaml:"path,omitempty"`
	ReadOnly  bool   `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	SizeBytes int64  `json:"sizeBytes,omitempty" yaml:"sizeBytes,omitempty"`
}
