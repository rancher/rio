package v1

import (
	"time"

	"github.com/rancher/wrangler/pkg/condition"
	"github.com/rancher/wrangler/pkg/genericcondition"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ServiceConditionCurrentRevision  = condition.Cond("CurrentRevision")
	ServiceConditionImageReady       = condition.Cond("ImageReady")
	ServiceConditionDeploymentStable = condition.Cond("DeploymentStable")
	ServiceConditionPromoted         = condition.Cond("Promoted")
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Service struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceSpec   `json:"spec,omitempty"`
	Status ServiceStatus `json:"status,omitempty"`
}

type ServiceRevision struct {
	Version string `json:"version,omitempty"`
	Weight  int    `json:"weight,omitempty"`
	App     string `json:"app,omitempty"`
}

type ServiceScale struct {
	Scale           int `json:"scale,omitempty"`
	UpdateBatchSize int `json:"updateBatchSize,omitempty"`
}

type AutoscaleConfig struct {
	Concurrency int `json:"concurrency,omitempty"`
	MinScale    int `json:"minScale,omitempty"`
	MaxScale    int `json:"maxScale,omitempty"`
}

type SystemServiceSpec struct {
	UpdateOrder        string                     `json:"updateOrder,omitempty"`
	UpdateStrategy     string                     `json:"updateStrategy,omitempty"`
	DeploymentStrategy string                     `json:"deploymentStrategy,omitempty"`
	Global             bool                       `json:"global,omitempty"`
	VolumeTemplates    []v1.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty"`
	PodSpec            v1.PodSpec                 `json:"podSpec,omitempty"`
}

type RolloutConfig struct {
	Rollout          bool          `json:"rollout,omitempty"`
	RolloutIncrement int           `json:"rolloutIncrement,omitempty"`
	RolloutInterval  time.Duration `json:"rolloutInterval,omitempty"`
}

type ServiceSpec struct {
	ServiceScale
	ServiceRevision
	AutoscaleConfig
	RolloutConfig
	PodConfig

	DisableServiceMesh bool         `json:"disableServiceMesh,omitempty"`
	Permissions        []Permission `json:"permissions,omitempty"`
	GlobalPermissions  []Permission `json:"globalPermissions,omitempty"`
}

type PodDNSConfig struct {
	Nameservers []string             `json:"nameservers,omitempty"`
	Searches    []string             `json:"searches,omitempty"`
	Options     []PodDNSConfigOption `json:"options,omitempty"`
}

type PodDNSConfigOption struct {
	Name  string  `json:"name,omitempty"`
	Value *string `json:"value,omitempty"`
}

type ContainerSecurityContext struct {
	RunAsUser              *int64 `json:"runAsUser,omitempty"`
	RunAsGroup             *int64 `json:"runAsGroup,omitempty"`
	ReadOnlyRootFilesystem *bool  `json:"readOnlyRootFilesystem,omitempty"`
}

type NamedContainer struct {
	Name string `json:"name,omitempty"`
	Init bool   `json:"init,omitempty"`
	Container
}

type Container struct {
	Image           string            `json:"image,omitempty"`
	Build           ImageBuild        `json:"build,omitempty"`
	Command         []string          `json:"command,omitempty"`
	Args            []string          `json:"args,omitempty"`
	WorkingDir      string            `json:"workingDir,omitempty"`
	Ports           []ContainerPort   `json:"ports,omitempty"`
	Env             []EnvVar          `json:"env,omitempty"`
	CPUs            resource.Quantity `json:"cpus,omitempty"`
	Memory          resource.Quantity `json:"memory,omitempty"`
	Secrets         []DataMount       `json:"secrets,omitempty"`
	Configs         []DataMount       `json:"configs,omitempty"`
	LivenessProbe   *v1.Probe         `json:"livenessProbe,omitempty"`
	ReadinessProbe  *v1.Probe         `json:"readinessProbe,omitempty"`
	ImagePullPolicy v1.PullPolicy     `json:"imagePullPolicy,omitempty"`
	Stdin           bool              `json:"stdin,omitempty"`
	StdinOnce       bool              `json:"stdinOnce,omitempty"`
	TTY             bool              `json:"tty,omitempty"`

	ContainerSecurityContext
}

type DataMount struct {
	Directory string `json:"directory,omitempty"`
	Name      string `json:"name,omitempty"`
	File      string `json:"file,omitempty"`
	Key       string `json:"key,omitempty"`
}

type EnvVar struct {
	Name          string `json:"name,omitempty"`
	Value         string `json:"value,omitempty"`
	SecretName    string `json:"secretName,omitempty"`
	ConfigMapName string `json:"configMapName,omitempty"`
	Key           string `json:"key,omitempty"`
	Optional      *bool  `json:"optional,omitempty"`
}

type PodConfig struct {
	Sidecars    []NamedContainer `json:"containers,omitempty"`
	DNSPolicy   v1.DNSPolicy     `json:"dnsPolicy,omitempty"`
	Hostname    string           `json:"hostname,omitempty"`
	HostAliases []v1.HostAlias   `json:"hostAliases,omitempty"`

	PodDNSConfig
	Container
}

type Protocol string

const (
	ProtocolTCP   Protocol = "TCP"
	ProtocolUDP   Protocol = "UDP"
	ProtocolSCTP  Protocol = "SCTP"
	ProtocolHTTP  Protocol = "HTTP"
	ProtocolHTTP2 Protocol = "HTTP2"
	ProtocolGRPC  Protocol = "GRPC"
)

type ContainerPort struct {
	Name         string   `json:"name,omitempty"`
	InternalOnly bool     `json:"internalOnly,omitempty"`
	Protocol     Protocol `json:"protocol,omitempty"`
	Port         int32    `json:"port"`
	TargetPort   int32    `json:"targetPort,omitempty"`
}

type ServiceStatus struct {
	DeploymentStatus  *appsv1.DeploymentStatus            `json:"deploymentStatus,omitempty"`
	DaemonSetStatus   *appsv1.DaemonSetStatus             `json:"daemonSetStatus,omitempty"`
	StatefulSetStatus *appsv1.StatefulSetStatus           `json:"statefulSetStatus,omitempty"`
	ScaleStatus       *ScaleStatus                        `json:"scaleStatus,omitempty"`
	ObservedScale     *int                                `json:"observedScale,omitempty"`
	ScaleOverride     *int                                `json:"scaleOverride,omitempty"`
	ObservedWeight    *int                                `json:"observedWeight,omitempty"`
	WeightOverride    *int                                `json:"weightOverride,omitempty"`
	ContainerImages   map[string]string                   `json:"containerImages,omitempty"`
	Conditions        []genericcondition.GenericCondition `json:"conditions,omitempty"`
	Endpoints         []Endpoint                          `json:"endpoints,omitempty"`
	PublicDomains     []string                            `json:"publicDomains,omitempty"`
}

type Endpoint struct {
	URL string `json:"url,omitempty"`
}

type ScaleStatus struct {
	Ready       int `json:"ready,omitempty"`
	Unavailable int `json:"unavailable,omitempty"`
	Available   int `json:"available,omitempty"`
	Updated     int `json:"updated,omitempty"`
}

type ImageBuild struct {
	Repo       string `json:"repo,omitempty"`
	Tag        string `json:"tag,omitempty"`
	Commit     string `json:"commit,omitempty"`
	Branch     string `json:"branch,omitempty"`
	StageOnly  string `json:"stageOnly,omitempty"`
	DockerFile string `json:"dockerFile,omitempty"`
	Template   string `json:"template,omitempty"`
	Secret     string `json:"secret,omitempty"`
	Hook       bool   `json:"hook,omitempty"`
}
