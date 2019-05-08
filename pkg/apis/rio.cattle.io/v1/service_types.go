package v1

import (
	"bytes"
	"strconv"

	"github.com/rancher/rio/pkg/apis/common"
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

// Service acts as a top level resource for a container and its sidecarsm and routing resources.
// Each service represents an individual revision, group by Spec.App(defaults to Service.Name), and Spec.Version(defaults to v0)
type Service struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceSpec   `json:"spec,omitempty"`
	Status ServiceStatus `json:"status,omitempty"`
}

// ServiceRevision speficies the APP name, Version and Weight to uniquely identify each Revision
type ServiceRevision struct {
	// Revision Version
	Version string `json:"version,omitempty"`

	// Revision Weight
	Weight int `json:"weight,omitempty"`

	// Revision App name
	App string `json:"app,omitempty"`
}

// ServiceScale Specifies the scale parameters for Service
type ServiceScale struct {
	// Number of desired pods. This is a pointer to distinguish between explicit zero and not specified. Defaults to 1.
	Scale int `json:"scale,omitempty"`

	// The maximum number of pods that can be scheduled above the desired number of pods. Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// This can not be 0 if MaxUnavailable is 0. Absolute number is calculated from percentage by rounding up.
	// Defaults to 25%. Example: when this is set to 30%, the new ReplicaSet can be scaled up immediately when the rolling update starts, such that the total number of old and new pods do not exceed 130% of desired pods.
	// Once old pods have been killed, new ReplicaSet can be scaled up further, ensuring that total number of pods running at any time during the update is at most 130% of desired pods.
	//
	// +optional
	UpdateBatchSize int `json:"updateBatchSize,omitempty"`
}

type AutoscaleConfig struct {
	// ContainerConcurrency specifies the maximum allowed in-flight (concurrent) requests per container of the Revision. Defaults to 0 which means unlimited concurrency.
	// This field replaces ConcurrencyModel. A value of 1 is equivalent to Single and 0 is equivalent to Multi.
	Concurrency *int `json:"concurrency,omitempty"`

	// The minimal scale Service can be scaled
	MinScale *int `json:"minScale,omitempty"`

	// The maximum scale Service can be scaled
	MaxScale *int `json:"maxScale,omitempty"`
}

type SystemServiceSpec struct {
	UpdateOrder        string                     `json:"updateOrder,omitempty"`
	UpdateStrategy     string                     `json:"updateStrategy,omitempty"`
	DeploymentStrategy string                     `json:"deploymentStrategy,omitempty"`
	Global             bool                       `json:"global,omitempty"`
	VolumeTemplates    []v1.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty"`
	PodSpec            v1.PodSpec                 `json:"podSpec,omitempty"`
}

// RolloutConfig specifies the configuration when promoting a new revision
type RolloutConfig struct {
	// Whether to turn on Rollout(changing the weight gradually)
	Rollout bool `json:"rollout,omitempty"`

	// Increment Value each Rollout can scale up or down
	RolloutIncrement int `json:"rolloutIncrement,omitempty"`

	// Increment Interval between each Rollout
	RolloutInterval int `json:"rolloutInterval,omitempty"`
}

// ServiceSpec represents spec for Service
type ServiceSpec struct {
	ServiceScale
	ServiceRevision
	AutoscaleConfig
	RolloutConfig
	PodConfig

	// Whether to disable ServiceMesh for Service. If true, no mesh sidecar will be deployed along with the Service
	DisableServiceMesh bool `json:"disableServiceMesh,omitempty"`

	// Permissions to the Services. It will create corresponding ServiceAccounts, Roles and RoleBinding.
	Permissions []Permission `json:"permissions,omitempty"`

	// GlobalPermissions to the Services. It will create corresponding ServiceAccounts, ClusterRoles and ClusterRoleBinding.
	GlobalPermissions []Permission `json:"globalPermissions,omitempty"`

	// System Field Spec
	SystemSpec *SystemServiceSpec `json:"systemSpec,omitempty"`
}

// PodDNSConfig Specifies the DNS parameters of a pod. Parameters specified here will be merged to the generated DNS configuration based on DNSPolicy.
type PodDNSConfig struct {
	// A list of DNS name server IP addresses. This will be appended to the base nameservers generated from DNSPolicy. Duplicated nameservers will be removed.
	Nameservers []string `json:"dnsNameservers,omitempty"`

	// A list of DNS search domains for host-name lookup. This will be appended to the base search paths generated from DNSPolicy. Duplicated search paths will be removed.
	Searches []string `json:"dnsSearches,omitempty"`

	// A list of DNS resolver options. This will be merged with the base options generated from DNSPolicy.
	// Duplicated entries will be removed. Resolution options given in Options will override those that appear in the base DNSPolicy.
	Options []PodDNSConfigOption `json:"dnsOptions,omitempty"`
}

type PodDNSConfigOption struct {
	Name  string  `json:"name,omitempty"`
	Value *string `json:"value,omitempty"`
}

// ContainerSecurityContext holds pod-level security attributes and common container settings. Optional: Defaults to empty. See type description for default values of each field.
type ContainerSecurityContext struct {
	// The UID to run the entrypoint of the container process. Defaults to user specified in image metadata if unspecified. May also be set in SecurityContext.
	// If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence for that container
	RunAsUser *int64 `json:"runAsUser,omitempty"`

	// The GID to run the entrypoint of the container process. Uses runtime default if unset. May also be set in SecurityContext.
	// If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence for that container.
	RunAsGroup *int64 `json:"runAsGroup,omitempty"`

	// Whether this container has a read-only root filesystem. Default is false.
	ReadOnlyRootFilesystem *bool `json:"readOnlyRootFilesystem,omitempty"`
}

type NamedContainer struct {
	// The name of the container
	Name string `json:"name,omitempty"`

	// List of initialization containers belonging to the pod.
	// Init containers are executed in order prior to containers being started.
	// If any init container fails, the pod is considered to have failed and is handled according to its restartPolicy.
	// The name for an init container or normal container must be unique among all containers.
	// Init containers may not have Lifecycle actions, Readiness probes, or Liveness probes.
	// The resourceRequirements of an init container are taken into account during scheduling by finding the highest request/limit for each resource type, and then using the max of of that value or the sum of the normal containers.
	// Limits are applied to init containers in a similar fashion. Init containers cannot currently be added or removed. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
	Init bool `json:"init,omitempty"`

	Container
}

type Container struct {
	// Docker image name. More info: https://kubernetes.io/docs/concepts/containers/images This field is optional to allow higher level config management to default or override container images in workload controllers like Deployments and StatefulSets.
	Image string `json:"image,omitempty"`

	// ImageBuild Specify the build parameter
	Build *ImageBuild `json:"build,omitempty"`

	// Entrypoint array. Not executed within a shell. The docker image's ENTRYPOINT is used if this is not provided.
	// Variable references $(VAR_NAME) are expanded using the container's environment. If a variable cannot be resolved, the reference in the input string will be unchanged.
	// The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not.
	// Cannot be updated. More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell
	Command []string `json:"command,omitempty"`

	// Arguments to the entrypoint. The docker image's CMD is used if this is not provided.
	// Variable references $(VAR_NAME) are expanded using the container's environment.
	// If a variable cannot be resolved, the reference in the input string will be unchanged.
	// The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not.
	// Cannot be updated. More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell
	Args []string `json:"args,omitempty"`

	// Container's working directory. If not specified, the container runtime's default will be used, which might be configured in the container image. Cannot be updated.
	WorkingDir string `json:"workingDir,omitempty"`

	// List of ports to expose from the container. Exposing a port here gives the system additional information about the network connections a container uses, but is primarily informational. Not specifying a port here DOES NOT prevent that port from being exposed.
	// Any port which is listening on the default "0.0.0.0" address inside a container will be accessible from the network. Cannot be updated.
	Ports []ContainerPort `json:"ports,omitempty"`

	// List of environment variables to set in the container. Cannot be updated.
	Env []EnvVar `json:"env,omitempty"`

	// CPU, in cores. (500m = .5 cores)
	CPUs *resource.Quantity `json:"cpus,omitempty"`

	// Memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)
	Memory *resource.Quantity `json:"memory,omitempty"`

	// Secrets Mounts
	Secrets []DataMount `json:"secrets,omitempty"`

	// Configmaps Mounts
	Configs []DataMount `json:"configs,omitempty"`

	// Periodic probe of container liveness. Container will be restarted if the probe fails. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	LivenessProbe *v1.Probe `json:"livenessProbe,omitempty"`

	// Periodic probe of container service readiness. Container will be removed from service endpoints if the probe fails. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	ReadinessProbe *v1.Probe `json:"readinessProbe,omitempty"`

	// Image pull policy. One of Always, Never, IfNotPresent. Defaults to Always if :latest tag is specified, or IfNotPresent otherwise. Cannot be updated. More info: https://kubernetes.io/docs/concepts/containers/images#updating-images
	ImagePullPolicy v1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// Whether this container should allocate a buffer for stdin in the container runtime. If this is not set, reads from stdin in the container will always result in EOF. Default is false.
	Stdin bool `json:"stdin,omitempty"`

	// Whether the container runtime should close the stdin channel after it has been opened by a single attach. When stdin is true the stdin stream will remain open across multiple attach sessions.
	// If stdinOnce is set to true, stdin is opened on container start, is empty until the first client attaches to stdin, and then remains open and accepts data until the client disconnects, at which time stdin is closed and remains closed until the container is restarted. If this flag is false, a container processes that reads from stdin will never receive an EOF. Default is false
	StdinOnce bool `json:"stdinOnce,omitempty"`

	// Whether this container should allocate a TTY for itself, also requires 'stdin' to be true. Default is false.
	TTY bool `json:"tty,omitempty"`

	// Pod volumes to mount into the container's filesystem. Cannot be updated.
	Volumes []Volume `json:"volumes,omitempty"`

	ContainerSecurityContext
}

type DataMount struct {
	Directory string `json:"directory,omitempty"`
	Name      string `json:"name,omitempty"`
	File      string `json:"file,omitempty"`
	Key       string `json:"key,omitempty"`
}

type Volume struct {
	Name string
	Path string
}

type EnvVar struct {
	Name          string `json:"name,omitempty"`
	Value         string `json:"value,omitempty"`
	SecretName    string `json:"secretName,omitempty"`
	ConfigMapName string `json:"configMapName,omitempty"`
	Key           string `json:"key,omitempty"`
}

type PodConfig struct {
	// List of containers belonging to the pod. Containers cannot currently be added or removed. There must be at least one container in a Pod. Cannot be updated.
	Sidecars []NamedContainer `json:"containers,omitempty"`

	// Set DNS policy for the pod. Defaults to "ClusterFirst". Valid values are 'ClusterFirstWithHostNet', 'ClusterFirst', 'Default' or 'None'.
	// DNS parameters given in DNSConfig will be merged with the policy selected with DNSPolicy.
	// To have DNS options set along with hostNetwork, you have to specify DNS policy explicitly to 'ClusterFirstWithHostNet'.
	DNSPolicy v1.DNSPolicy `json:"dnsPolicy,omitempty"`

	// Specifies the hostname of the Pod If not specified, the pod's hostname will be set to a system-defined value.
	Hostname string `json:"hostname,omitempty"`

	// HostAliases is an optional list of hosts and IPs that will be injected into the pod's hosts file if specified. This is only valid for non-hostNetwork pods.
	HostAliases []v1.HostAlias `json:"hostAliases,omitempty"`

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

func (c ContainerPort) MaybeString() interface{} {
	b := bytes.Buffer{}
	if c.Port != 0 && c.TargetPort != 0 {
		b.WriteString(strconv.FormatInt(int64(c.Port), 10))
		b.WriteString(":")
		b.WriteString(strconv.FormatInt(int64(c.TargetPort), 10))
	} else if c.TargetPort != 0 {
		b.WriteString(strconv.FormatInt(int64(c.TargetPort), 10))
	}

	if b.Len() > 0 && c.Protocol != "" && c.Protocol != "tcp" {
		b.WriteString("/")
		b.WriteString(string(c.Protocol))
	}

	return b.String()
}

type ServiceStatus struct {
	// Most recently observed status of the Deployment.
	DeploymentStatus *appsv1.DeploymentStatus `json:"deploymentStatus,omitempty"`

	// ScaleStatus for the Service
	ScaleStatus *ScaleStatus `json:"scaleStatus,omitempty"`

	// Last timestamp scaled from zero replica
	ScaleFromZeroTimestamp *metav1.Time `json:"scaleFromZeroTimestamp,omitempty"`

	// ObservedScale is calcaluted from autoscaling component to make sure pod has the desired load
	ObservedScale *int `json:"observedScale,omitempty"`

	// WeightOverride is the weight calculated from serviceset revision
	WeightOverride *int `json:"weightOverride,omitempty"`

	ContainerImages map[string]string `json:"containerImages,omitempty"`

	// Represents the latest available observations of a deployment's current state.
	Conditions []genericcondition.GenericCondition `json:"conditions,omitempty"`

	// The Endpoints to access the service
	Endpoints []string `json:"endpoints,omitempty"`

	// The list of publicdomains pointing to the service
	PublicDomains []string `json:"publicDomains,omitempty"`
}

type ScaleStatus struct {
	// Total number of ready pods targeted by this deployment.
	Ready int `json:"ready,omitempty"`

	// Total number of unavailable pods targeted by this deployment. This is the total number of pods that are still required for the deployment to have 100% available capacity.
	// They may either be pods that are running but not yet available or pods that still have not been created.
	Unavailable int `json:"unavailable,omitempty"`

	// Total number of available pods (ready for at least minReadySeconds) targeted by this deployment.
	Available int `json:"available,omitempty"`

	// Total number of non-terminated pods targeted by this deployment that have the desired template spec.
	Updated int `json:"updated,omitempty"`
}

type ImageBuild struct {
	// Repository url
	Repo string `json:"repo,omitempty"`

	// Repo Revision. Can be a git commit or tag
	Revision string `json:"revision,omitempty"`

	// Repo Branch. If specified, a gitmodule will be created to watch the repo and creating new revision if new commit or tag is pushed.
	Branch string `json:"branch,omitempty"`

	// Whether to only stage the new revision. If true, the new created service will not be allocating any traffic automatically.
	StageOnly bool `json:"stageOnly,omitempty"`

	// Specify the name Of the Dockerfile in the Repo. Defaults to `Dockerfile`.
	DockerFile string `json:"dockerFile,omitempty"`

	// Specify the build template. Defaults to `buildkit`.
	Template string `json:"template,omitempty"`

	// Specify the secret name. If specified, it will register a webhook and only creates new revision if webhook is triggered.
	Secret string `json:"secret,omitempty"`
}

func (in *Service) State() common.State {
	state := common.StateFromConditionAndMeta(in.ObjectMeta, in.Status.Conditions)
	if len(in.Status.Conditions) == 0 {
		state.State = "pending"
	}
	if scaleIsZero(in) {
		state.State = "inactive"
	}
	return state
}

func scaleIsZero(service *Service) bool {
	if service.Status.ScaleStatus == nil {
		return true
	}
	ready := service.Status.ScaleStatus.Ready
	available := service.Status.ScaleStatus.Available
	unavailable := service.Status.ScaleStatus.Unavailable
	updated := service.Status.ScaleStatus.Updated
	scale := service.Spec.Scale

	return ready+available+unavailable+updated+scale == 0
}
