package client

const (
	SecurityContextType                          = "securityContext"
	SecurityContextFieldAllowPrivilegeEscalation = "allowPrivilegeEscalation"
	SecurityContextFieldCapabilities             = "capabilities"
	SecurityContextFieldPrivileged               = "privileged"
	SecurityContextFieldProcMount                = "procMount"
	SecurityContextFieldReadOnlyRootFilesystem   = "readOnlyRootFilesystem"
	SecurityContextFieldRunAsGroup               = "runAsGroup"
	SecurityContextFieldRunAsNonRoot             = "runAsNonRoot"
	SecurityContextFieldRunAsUser                = "runAsUser"
	SecurityContextFieldSELinuxOptions           = "seLinuxOptions"
)

type SecurityContext struct {
	AllowPrivilegeEscalation *bool           `json:"allowPrivilegeEscalation,omitempty" yaml:"allowPrivilegeEscalation,omitempty"`
	Capabilities             *Capabilities   `json:"capabilities,omitempty" yaml:"capabilities,omitempty"`
	Privileged               *bool           `json:"privileged,omitempty" yaml:"privileged,omitempty"`
	ProcMount                string          `json:"procMount,omitempty" yaml:"procMount,omitempty"`
	ReadOnlyRootFilesystem   *bool           `json:"readOnlyRootFilesystem,omitempty" yaml:"readOnlyRootFilesystem,omitempty"`
	RunAsGroup               *int64          `json:"runAsGroup,omitempty" yaml:"runAsGroup,omitempty"`
	RunAsNonRoot             *bool           `json:"runAsNonRoot,omitempty" yaml:"runAsNonRoot,omitempty"`
	RunAsUser                *int64          `json:"runAsUser,omitempty" yaml:"runAsUser,omitempty"`
	SELinuxOptions           *SELinuxOptions `json:"seLinuxOptions,omitempty" yaml:"seLinuxOptions,omitempty"`
}
