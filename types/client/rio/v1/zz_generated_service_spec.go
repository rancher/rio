package client

const (
	ServiceSpecType                        = "serviceSpec"
	ServiceSpecFieldAutoScale              = "autoScale"
	ServiceSpecFieldBatchSize              = "batchSize"
	ServiceSpecFieldCPUs                   = "nanoCpus"
	ServiceSpecFieldCapAdd                 = "capAdd"
	ServiceSpecFieldCapDrop                = "capDrop"
	ServiceSpecFieldCommand                = "command"
	ServiceSpecFieldConfigs                = "configs"
	ServiceSpecFieldDNS                    = "dns"
	ServiceSpecFieldDNSOptions             = "dnsOptions"
	ServiceSpecFieldDNSSearch              = "dnsSearch"
	ServiceSpecFieldDefaultVolumeDriver    = "defaultVolumeDriver"
	ServiceSpecFieldDeploymentStrategy     = "deploymentStrategy"
	ServiceSpecFieldDevices                = "devices"
	ServiceSpecFieldEntrypoint             = "entrypoint"
	ServiceSpecFieldEnvironment            = "environment"
	ServiceSpecFieldExposedPorts           = "expose"
	ServiceSpecFieldExtraHosts             = "extraHosts"
	ServiceSpecFieldGlobal                 = "global"
	ServiceSpecFieldGlobalPermissions      = "globalPermissions"
	ServiceSpecFieldGroup                  = "group"
	ServiceSpecFieldHealthcheck            = "healthcheck"
	ServiceSpecFieldHostname               = "hostname"
	ServiceSpecFieldImage                  = "image"
	ServiceSpecFieldImageBuild             = "build"
	ServiceSpecFieldImagePullPolicy        = "imagePullPolicy"
	ServiceSpecFieldInit                   = "init"
	ServiceSpecFieldIpcMode                = "ipc"
	ServiceSpecFieldMemoryLimitBytes       = "memoryLimitBytes"
	ServiceSpecFieldMemoryReservationBytes = "memoryReservationBytes"
	ServiceSpecFieldMetadata               = "metadata"
	ServiceSpecFieldNetworkMode            = "net"
	ServiceSpecFieldOpenStdin              = "stdinOpen"
	ServiceSpecFieldParentService          = "parentService"
	ServiceSpecFieldPermissions            = "permissions"
	ServiceSpecFieldPidMode                = "pid"
	ServiceSpecFieldPortBindings           = "ports"
	ServiceSpecFieldPrivileged             = "privileged"
	ServiceSpecFieldProjectID              = "projectId"
	ServiceSpecFieldPromote                = "promote"
	ServiceSpecFieldReadonlyRootfs         = "readOnly"
	ServiceSpecFieldReadycheck             = "readycheck"
	ServiceSpecFieldRestartPolicy          = "restart"
	ServiceSpecFieldScale                  = "scale"
	ServiceSpecFieldScheduling             = "scheduling"
	ServiceSpecFieldSecrets                = "secrets"
	ServiceSpecFieldServiceLabels          = "serviceLabels"
	ServiceSpecFieldSidekicks              = "sidekicks"
	ServiceSpecFieldStackID                = "stackId"
	ServiceSpecFieldStopGracePeriodSeconds = "stopGracePeriod"
	ServiceSpecFieldTmpfs                  = "tmpfs"
	ServiceSpecFieldTty                    = "tty"
	ServiceSpecFieldUpdateOrder            = "updateOrder"
	ServiceSpecFieldUpdateStrategy         = "updateStrategy"
	ServiceSpecFieldUser                   = "user"
	ServiceSpecFieldVersion                = "version"
	ServiceSpecFieldVolumes                = "volumes"
	ServiceSpecFieldVolumesFrom            = "volumesFrom"
	ServiceSpecFieldWeight                 = "weight"
	ServiceSpecFieldWorkingDir             = "workingDir"
)

type ServiceSpec struct {
	AutoScale              *AutoscaleConfig          `json:"autoScale,omitempty" yaml:"autoScale,omitempty"`
	BatchSize              int64                     `json:"batchSize,omitempty" yaml:"batchSize,omitempty"`
	CPUs                   string                    `json:"nanoCpus,omitempty" yaml:"nanoCpus,omitempty"`
	CapAdd                 []string                  `json:"capAdd,omitempty" yaml:"capAdd,omitempty"`
	CapDrop                []string                  `json:"capDrop,omitempty" yaml:"capDrop,omitempty"`
	Command                []string                  `json:"command,omitempty" yaml:"command,omitempty"`
	Configs                []ConfigMapping           `json:"configs,omitempty" yaml:"configs,omitempty"`
	DNS                    []string                  `json:"dns,omitempty" yaml:"dns,omitempty"`
	DNSOptions             []string                  `json:"dnsOptions,omitempty" yaml:"dnsOptions,omitempty"`
	DNSSearch              []string                  `json:"dnsSearch,omitempty" yaml:"dnsSearch,omitempty"`
	DefaultVolumeDriver    string                    `json:"defaultVolumeDriver,omitempty" yaml:"defaultVolumeDriver,omitempty"`
	DeploymentStrategy     string                    `json:"deploymentStrategy,omitempty" yaml:"deploymentStrategy,omitempty"`
	Devices                []DeviceMapping           `json:"devices,omitempty" yaml:"devices,omitempty"`
	Entrypoint             []string                  `json:"entrypoint,omitempty" yaml:"entrypoint,omitempty"`
	Environment            []string                  `json:"environment,omitempty" yaml:"environment,omitempty"`
	ExposedPorts           []ExposedPort             `json:"expose,omitempty" yaml:"expose,omitempty"`
	ExtraHosts             []string                  `json:"extraHosts,omitempty" yaml:"extraHosts,omitempty"`
	Global                 bool                      `json:"global,omitempty" yaml:"global,omitempty"`
	GlobalPermissions      []Permission              `json:"globalPermissions,omitempty" yaml:"globalPermissions,omitempty"`
	Group                  string                    `json:"group,omitempty" yaml:"group,omitempty"`
	Healthcheck            *HealthConfig             `json:"healthcheck,omitempty" yaml:"healthcheck,omitempty"`
	Hostname               string                    `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	Image                  string                    `json:"image,omitempty" yaml:"image,omitempty"`
	ImageBuild             *ImageBuild               `json:"build,omitempty" yaml:"build,omitempty"`
	ImagePullPolicy        string                    `json:"imagePullPolicy,omitempty" yaml:"imagePullPolicy,omitempty"`
	Init                   bool                      `json:"init,omitempty" yaml:"init,omitempty"`
	IpcMode                string                    `json:"ipc,omitempty" yaml:"ipc,omitempty"`
	MemoryLimitBytes       int64                     `json:"memoryLimitBytes,omitempty" yaml:"memoryLimitBytes,omitempty"`
	MemoryReservationBytes int64                     `json:"memoryReservationBytes,omitempty" yaml:"memoryReservationBytes,omitempty"`
	Metadata               map[string]interface{}    `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	NetworkMode            string                    `json:"net,omitempty" yaml:"net,omitempty"`
	OpenStdin              bool                      `json:"stdinOpen,omitempty" yaml:"stdinOpen,omitempty"`
	ParentService          string                    `json:"parentService,omitempty" yaml:"parentService,omitempty"`
	Permissions            []Permission              `json:"permissions,omitempty" yaml:"permissions,omitempty"`
	PidMode                string                    `json:"pid,omitempty" yaml:"pid,omitempty"`
	PortBindings           []PortBinding             `json:"ports,omitempty" yaml:"ports,omitempty"`
	Privileged             bool                      `json:"privileged,omitempty" yaml:"privileged,omitempty"`
	ProjectID              string                    `json:"projectId,omitempty" yaml:"projectId,omitempty"`
	Promote                bool                      `json:"promote,omitempty" yaml:"promote,omitempty"`
	ReadonlyRootfs         bool                      `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	Readycheck             *HealthConfig             `json:"readycheck,omitempty" yaml:"readycheck,omitempty"`
	RestartPolicy          string                    `json:"restart,omitempty" yaml:"restart,omitempty"`
	Scale                  int64                     `json:"scale,omitempty" yaml:"scale,omitempty"`
	Scheduling             *Scheduling               `json:"scheduling,omitempty" yaml:"scheduling,omitempty"`
	Secrets                []SecretMapping           `json:"secrets,omitempty" yaml:"secrets,omitempty"`
	ServiceLabels          map[string]string         `json:"serviceLabels,omitempty" yaml:"serviceLabels,omitempty"`
	Sidekicks              map[string]SidekickConfig `json:"sidekicks,omitempty" yaml:"sidekicks,omitempty"`
	StackID                string                    `json:"stackId,omitempty" yaml:"stackId,omitempty"`
	StopGracePeriodSeconds *int64                    `json:"stopGracePeriod,omitempty" yaml:"stopGracePeriod,omitempty"`
	Tmpfs                  []Tmpfs                   `json:"tmpfs,omitempty" yaml:"tmpfs,omitempty"`
	Tty                    bool                      `json:"tty,omitempty" yaml:"tty,omitempty"`
	UpdateOrder            string                    `json:"updateOrder,omitempty" yaml:"updateOrder,omitempty"`
	UpdateStrategy         string                    `json:"updateStrategy,omitempty" yaml:"updateStrategy,omitempty"`
	User                   string                    `json:"user,omitempty" yaml:"user,omitempty"`
	Version                string                    `json:"version,omitempty" yaml:"version,omitempty"`
	Volumes                []Mount                   `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	VolumesFrom            []string                  `json:"volumesFrom,omitempty" yaml:"volumesFrom,omitempty"`
	Weight                 int64                     `json:"weight,omitempty" yaml:"weight,omitempty"`
	WorkingDir             string                    `json:"workingDir,omitempty" yaml:"workingDir,omitempty"`
}
