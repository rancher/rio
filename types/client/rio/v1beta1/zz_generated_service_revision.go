package client

const (
	ServiceRevisionType               = "serviceRevision"
	ServiceRevisionFieldParentService = "parentService"
	ServiceRevisionFieldPromote       = "promote"
	ServiceRevisionFieldVersion       = "version"
	ServiceRevisionFieldWeight        = "weight"
)

type ServiceRevision struct {
	ParentService string `json:"parentService,omitempty" yaml:"parentService,omitempty"`
	Promote       bool   `json:"promote,omitempty" yaml:"promote,omitempty"`
	Version       string `json:"version,omitempty" yaml:"version,omitempty"`
	Weight        int64  `json:"weight,omitempty" yaml:"weight,omitempty"`
}
