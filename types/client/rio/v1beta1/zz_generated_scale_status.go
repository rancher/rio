package client

const (
	ScaleStatusType             = "scaleStatus"
	ScaleStatusFieldAvailable   = "available"
	ScaleStatusFieldReady       = "ready"
	ScaleStatusFieldUnavailable = "unavailable"
	ScaleStatusFieldUpdated     = "updated"
)

type ScaleStatus struct {
	Available   int64 `json:"available,omitempty" yaml:"available,omitempty"`
	Ready       int64 `json:"ready,omitempty" yaml:"ready,omitempty"`
	Unavailable int64 `json:"unavailable,omitempty" yaml:"unavailable,omitempty"`
	Updated     int64 `json:"updated,omitempty" yaml:"updated,omitempty"`
}
