package client

const (
	DeviceMappingType             = "deviceMapping"
	DeviceMappingFieldInContainer = "inContainer"
	DeviceMappingFieldOnHost      = "onHost"
	DeviceMappingFieldPermissions = "permissions"
)

type DeviceMapping struct {
	InContainer string `json:"inContainer,omitempty" yaml:"inContainer,omitempty"`
	OnHost      string `json:"onHost,omitempty" yaml:"onHost,omitempty"`
	Permissions string `json:"permissions,omitempty" yaml:"permissions,omitempty"`
}
