package v1beta1

import (
	"bytes"
	"strconv"

	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/norman/types/values"
	"github.com/rancher/types/mapper"
	"k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Service struct {
	types.Namespaced

	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceSpec   `json:"spec,omitempty"`
	Status ServiceStatus `json:"status,omitempty"`
}

type ServiceRevision struct {
	Spec    ServiceUnversionedSpec `json:"spec,omitempty"`
	Weight  int                    `json:"weight,omitempty"`
	Promote bool                   `json:"promote,omitempty"`
	Status  ServiceStatus          `json:"status,omitempty"`
}

type ServiceUnversionedSpec struct {
	Labels             map[string]string `json:"labels,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"` //alias annotations
	Scale              int               `json:"scale"`
	BatchSize          int               `json:"batchSize,omitempty"`
	UpdateOrder        string            `json:"updateOrder,omitempty" norman:"type=enum,options=start-first|stop-first"`
	UpdateStrategy     string            `json:"updateStrategy,omitempty" norman:"type=enum,options=rolling|on-delete,default=rolling"`
	DeploymentStrategy string            `json:"deploymentStrategy,omitempty" norman:"type=enum,options=parallel|ordered,default=parallel"`

	PodConfig
	PrivilegedConfig
	Sidekicks map[string]SidekickConfig `json:"sidekicks,omitempty"`

	ContainerConfig
}

type ServiceSpec struct {
	ServiceUnversionedSpec
	StackScoped
	Revisions map[string]ServiceRevision `json:"revisions,omitempty"`
}

type ServiceStatus struct {
	DeploymentStatus *v1beta2.DeploymentStatus `json:"deploymentStatus,omitempty"`
	ScaleStatus      *ScaleStatus              `json:"scaleStatus,omitempty"`
	Conditions       []Condition               `json:"conditions,omitempty"`
}

type ScaleStatus struct {
	Ready       int `json:"ready,omitempty"`
	Unavailable int `json:"unavailable,omitempty"`
	Available   int `json:"available,omitempty"`
	Updated     int `json:"updated,omitempty"`
}

type PodConfig struct {
	Hostname               string        `json:"hostname,omitempty"`
	Global                 bool          `json:"global,omitempty"`
	Scheduling             Scheduling    `json:"scheduling,omitempty"`
	StopGracePeriodSeconds *int          `json:"stopGracePeriod,omitempty"`                                                           // support friendly numbers
	RestartPolicy          string        `json:"restart,omitempty" norman:"type=enum,options=never|on-failure|always,default=always"` //support no and OnFailure
	DNS                    []string      `json:"dns,omitempty"`                                                                       // support string
	DNSOptions             []string      `json:"dnsOptions,omitempty"`                                                                // support string
	DNSSearch              []string      `json:"dnsSearch,omitempty"`                                                                 // support string
	ExtraHosts             []string      `json:"extraHosts,omitempty"`                                                                // support map
	GlobalPermissions      []Permission  `json:"globalPermissions,omitempty"`
	Permissions            []Permission  `json:"permissions,omitempty"`
	PortBindings           []PortBinding `json:"ports,omitempty"` // support []string
}

type Scheduling struct {
	Node      NodeScheduling `json:"node,omitempty"`
	Scheduler string         `json:"scheduler,omitempty"`
}

func (s Scheduling) ToNodeAffinity() (*v1.NodeAffinity, error) {
	data, err := convert.EncodeToMap(&s)
	if err != nil {
		return nil, err
	}
	mapper.SchedulingMapper{}.ToInternal(data)
	nodeAffinityMap, ok := values.GetValue(data, "affinity", "nodeAffinity")
	if !ok || convert.IsAPIObjectEmpty(nodeAffinityMap) {
		return nil, nil
	}
	nodeAffinity := &v1.NodeAffinity{}
	return nodeAffinity, convert.ToObj(nodeAffinityMap, nodeAffinity)
}

type NodeScheduling struct {
	NodeName   string   `json:"nodeName,omitempty" norman:"type=reference[/v1beta1/schemas/node]"`
	RequireAll []string `json:"requireAll,omitempty"`
	RequireAny []string `json:"requireAny,omitempty"`
	Preferred  []string `json:"preferred,omitempty"`
}

type PrivilegedConfig struct {
	NetworkMode string `json:"net,omitempty" norman:"type=enum,options=default|host,default=default"` // alias network, support bridge
	IpcMode     string `json:"ipc,omitempty" norman:"type=enum,options=default|host,default=default"`
	PidMode     string `json:"pid,omitempty" norman:"type=enum,options=default|host,default=default"`
}

type ContainerPrivilegedConfig struct {
	Privileged bool `json:"privileged,omitempty"`
}

type ExposedPort struct {
	Name string `json:"name,omitempty"`
	PortBinding
}

func (e ExposedPort) MaybeString() interface{} {
	s := e.PortBinding.MaybeString()
	if e.Name == "" {
		return s
	}
	return convert.ToString(s) + "," + e.Name
}

type PortBinding struct {
	Port       int64  `json:"port,omitempty"`
	Protocol   string `json:"protocol,omitempty"`
	IP         string `json:"ip,omitempty"`
	TargetPort int64  `json:"targetPort,omitempty"`
}

func (p PortBinding) MaybeString() interface{} {
	b := bytes.Buffer{}
	if p.Port != 0 && p.TargetPort != 0 {
		if p.IP != "" {
			b.WriteString(p.IP)
			b.WriteString(":")
		}
		b.WriteString(strconv.FormatInt(p.Port, 10))
		b.WriteString(":")
		b.WriteString(strconv.FormatInt(p.TargetPort, 10))
	} else if p.TargetPort != 0 {
		b.WriteString(strconv.FormatInt(p.TargetPort, 10))
	}

	if b.Len() > 0 && p.Protocol != "" && p.Protocol != "tcp" {
		b.WriteString("/")
		b.WriteString(p.Protocol)
	}

	return b.String()
}

type ContainerConfig struct {
	ContainerPrivilegedConfig

	CPUs                   string        `json:"nanoCpus,omitempty"`
	CapAdd                 []string      `json:"capAdd,omitempty"`  // support string
	CapDrop                []string      `json:"capDrop,omitempty"` // support string
	Command                []string      `json:"command,omitempty"` // support string
	DefaultVolumeDriver    string        `json:"defaultVolumeDriver,omitempty"`
	Entrypoint             []string      `json:"entrypoint,omitempty"`
	Environment            []string      `json:"environment,omitempty"` // alias env, support map
	ExposedPorts           []ExposedPort `json:"expose,omitempty"`      // support []string, map
	Healthcheck            *HealthConfig `json:"healthcheck,omitempty"`
	Image                  string        `json:"image,omitempty"`
	ImagePullPolicy        string        `json:"imagePullPolicy,omitempty" norman:"type=enum,options=always|never|not-present,default=not-present"`
	Init                   bool          `json:"init,omitempty"`
	MemoryLimitBytes       int64         `json:"memoryLimitBytes,omitempty"`
	MemoryReservationBytes int64         `json:"memoryReservationBytes,omitempty"`
	OpenStdin              bool          `json:"stdinOpen,omitempty"` // alias interactive
	ReadonlyRootfs         bool          `json:"readOnly,omitempty"`
	Tmpfs                  []Tmpfs       `json:"tmpfs,omitempty"` // support []string too
	Tty                    bool          `json:"tty,omitempty"`
	User                   string        `json:"user,omitempty"`
	Volumes                []Mount       `json:"volumes,omitempty"`     // support []string too
	VolumesFrom            []string      `json:"volumesFrom,omitempty"` // support []string too
	WorkingDir             string        `json:"workingDir,omitempty"`

	Devices []DeviceMapping `json:"devices,omitempty"` // support []string and map[string]string
	Configs []ConfigMapping `json:"configs,omitempty"`
	Secrets []SecretMapping `json:"secrets,omitempty"`
}

type SidekickConfig struct {
	InitContainer bool `json:"initContainer,omitempty"`
	ContainerConfig
}

type HealthConfig struct {
	// Test is the test to perform to check that the container is healthy.
	// An empty slice means to inherit the default.
	// The options are:
	// {} : inherit healthcheck
	// {"NONE"} : disable healthcheck
	// {"CMD", args...} : exec arguments directly
	// {"CMD-SHELL", command} : run command with system's default shell
	Test []string `json:"test,omitempty"` //alias string, deal with CMD, CMD-SHELL, NONE

	IntervalSeconds     int `json:"intervalSeconds,omitempty" norman:"default=10"`   // support friendly numbers, alias periodSeconds, period
	TimeoutSeconds      int `json:"timeoutSeconds,omitempty" norman:"default=5"`     // support friendly numbers
	InitialDelaySeconds int `json:"initialDelaySeconds,omitempty"`                   //alias start_period
	HealthyThreshold    int `json:"healthyThreshold,omitempty" norman:"default=2"`   //alias retries, successThreshold
	UnhealthyThreshold  int `json:"unhealthyThreshold,omitempty" norman:"default=3"` //alias failureThreshold, set to retries if unset
}

// DeviceMapping represents the device mapping between the host and the container.
type DeviceMapping struct {
	OnHost      string `json:"onHost,omitempty"`
	InContainer string `json:"inContainer,omitempty"`
	Permissions string `json:"permissions,omitempty"`
}

func (d DeviceMapping) MaybeString() interface{} {
	result := d.OnHost
	if len(d.InContainer) > 0 {
		if len(result) > 0 {
			result += ":"
		}
		result += d.InContainer
	}
	if len(d.Permissions) > 0 {
		if len(result) > 0 {
			result += ":"
		}
		result += d.Permissions
	}

	return result
}
