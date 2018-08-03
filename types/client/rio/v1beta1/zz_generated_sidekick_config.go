package client

const (
	SidekickConfigType                        = "sidekickConfig"
	SidekickConfigFieldCPUs                   = "nanoCpus"
	SidekickConfigFieldCapAdd                 = "capAdd"
	SidekickConfigFieldCapDrop                = "capDrop"
	SidekickConfigFieldCommand                = "command"
	SidekickConfigFieldConfigs                = "configs"
	SidekickConfigFieldDefaultVolumeDriver    = "defaultVolumeDriver"
	SidekickConfigFieldDevices                = "devices"
	SidekickConfigFieldEntrypoint             = "entrypoint"
	SidekickConfigFieldEnvironment            = "environment"
	SidekickConfigFieldExposedPorts           = "expose"
	SidekickConfigFieldHealthcheck            = "healthcheck"
	SidekickConfigFieldImage                  = "image"
	SidekickConfigFieldImagePullPolicy        = "imagePullPolicy"
	SidekickConfigFieldInit                   = "init"
	SidekickConfigFieldInitContainer          = "initContainer"
	SidekickConfigFieldMemoryLimitBytes       = "memoryLimitBytes"
	SidekickConfigFieldMemoryReservationBytes = "memoryReservationBytes"
	SidekickConfigFieldOpenStdin              = "stdinOpen"
	SidekickConfigFieldPrivileged             = "privileged"
	SidekickConfigFieldReadonlyRootfs         = "readOnly"
	SidekickConfigFieldSecrets                = "secrets"
	SidekickConfigFieldTmpfs                  = "tmpfs"
	SidekickConfigFieldTty                    = "tty"
	SidekickConfigFieldUser                   = "user"
	SidekickConfigFieldVolumes                = "volumes"
	SidekickConfigFieldVolumesFrom            = "volumesFrom"
	SidekickConfigFieldWorkingDir             = "workingDir"
)

type SidekickConfig struct {
	CPUs                   string          `json:"nanoCpus,omitempty" yaml:"nanoCpus,omitempty"`
	CapAdd                 []string        `json:"capAdd,omitempty" yaml:"capAdd,omitempty"`
	CapDrop                []string        `json:"capDrop,omitempty" yaml:"capDrop,omitempty"`
	Command                []string        `json:"command,omitempty" yaml:"command,omitempty"`
	Configs                []ConfigMapping `json:"configs,omitempty" yaml:"configs,omitempty"`
	DefaultVolumeDriver    string          `json:"defaultVolumeDriver,omitempty" yaml:"defaultVolumeDriver,omitempty"`
	Devices                []DeviceMapping `json:"devices,omitempty" yaml:"devices,omitempty"`
	Entrypoint             []string        `json:"entrypoint,omitempty" yaml:"entrypoint,omitempty"`
	Environment            []string        `json:"environment,omitempty" yaml:"environment,omitempty"`
	ExposedPorts           []ExposedPort   `json:"expose,omitempty" yaml:"expose,omitempty"`
	Healthcheck            *HealthConfig   `json:"healthcheck,omitempty" yaml:"healthcheck,omitempty"`
	Image                  string          `json:"image,omitempty" yaml:"image,omitempty"`
	ImagePullPolicy        string          `json:"imagePullPolicy,omitempty" yaml:"imagePullPolicy,omitempty"`
	Init                   bool            `json:"init,omitempty" yaml:"init,omitempty"`
	InitContainer          bool            `json:"initContainer,omitempty" yaml:"initContainer,omitempty"`
	MemoryLimitBytes       int64           `json:"memoryLimitBytes,omitempty" yaml:"memoryLimitBytes,omitempty"`
	MemoryReservationBytes int64           `json:"memoryReservationBytes,omitempty" yaml:"memoryReservationBytes,omitempty"`
	OpenStdin              bool            `json:"stdinOpen,omitempty" yaml:"stdinOpen,omitempty"`
	Privileged             bool            `json:"privileged,omitempty" yaml:"privileged,omitempty"`
	ReadonlyRootfs         bool            `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	Secrets                []SecretMapping `json:"secrets,omitempty" yaml:"secrets,omitempty"`
	Tmpfs                  []Tmpfs         `json:"tmpfs,omitempty" yaml:"tmpfs,omitempty"`
	Tty                    bool            `json:"tty,omitempty" yaml:"tty,omitempty"`
	User                   string          `json:"user,omitempty" yaml:"user,omitempty"`
	Volumes                []Mount         `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	VolumesFrom            []string        `json:"volumesFrom,omitempty" yaml:"volumesFrom,omitempty"`
	WorkingDir             string          `json:"workingDir,omitempty" yaml:"workingDir,omitempty"`
}
