package client

const (
	PodSecurityContextType                    = "podSecurityContext"
	PodSecurityContextFieldFSGroup            = "fsGroup"
	PodSecurityContextFieldRunAsGroup         = "runAsGroup"
	PodSecurityContextFieldRunAsNonRoot       = "runAsNonRoot"
	PodSecurityContextFieldRunAsUser          = "runAsUser"
	PodSecurityContextFieldSELinuxOptions     = "seLinuxOptions"
	PodSecurityContextFieldSupplementalGroups = "supplementalGroups"
)

type PodSecurityContext struct {
	FSGroup            *int64          `json:"fsGroup,omitempty" yaml:"fsGroup,omitempty"`
	RunAsGroup         *int64          `json:"runAsGroup,omitempty" yaml:"runAsGroup,omitempty"`
	RunAsNonRoot       *bool           `json:"runAsNonRoot,omitempty" yaml:"runAsNonRoot,omitempty"`
	RunAsUser          *int64          `json:"runAsUser,omitempty" yaml:"runAsUser,omitempty"`
	SELinuxOptions     *SELinuxOptions `json:"seLinuxOptions,omitempty" yaml:"seLinuxOptions,omitempty"`
	SupplementalGroups []int64         `json:"supplementalGroups,omitempty" yaml:"supplementalGroups,omitempty"`
}
