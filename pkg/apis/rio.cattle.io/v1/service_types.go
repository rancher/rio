package v1

import (
	"github.com/rancher/wrangler/pkg/condition"
	"github.com/rancher/wrangler/pkg/genericcondition"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	ServiceConditionImageReady      = condition.Cond("ImageReady")
	ServiceConditionServiceDeployed = condition.Cond("ServiceDeployed")
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Service acts as a top level resource for a container and its sidecars and routing resources.
// Each service represents an individual revision, group by Spec.App(defaults to Service.Name), and Spec.Version(defaults to v0)
type Service struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceSpec   `json:"spec,omitempty"`
	Status ServiceStatus `json:"status,omitempty"`
}

type AutoscaleConfig struct {
	// ContainerConcurrency specifies the maximum allowed in-flight (concurrent) requests per container of the Revision. Defaults to 0 which means unlimited concurrency.
	Concurrency int `json:"concurrency,omitempty"`

	// The minimal number of replicas Service can be scaled
	MinReplicas *int32 `json:"minReplicas,omitempty" mapper:"alias=minScale|min"`

	// The maximum number of replicas Service can be scaled
	MaxReplicas *int32 `json:"maxReplicas,omitempty" mapper:"alias=maxScale|max"`
}

// RolloutConfig specifies the configuration when promoting a new revision
type RolloutConfig struct {
	// Increment Value each Rollout can scale up or down
	Increment int `json:"increment,omitempty"`

	// Interval between each Rollout
	Interval metav1.Duration `json:"interval,omitempty" mapper:"duration"`

	// Pause if true the rollout will stop in place until set to false.
	Pause bool `json:"pause,omitempty"`
}

// ServiceSpec represents spec for Service
type ServiceSpec struct {
	PodConfig

	// Template this service is a template for new versions to be created base on changes
	// from the build.repo
	Template bool `json:"template,omitempty"`

	// StageOnly whether to only stage services that are generated through template from build.repo
	StageOnly bool `json:"stageOnly,omitempty"`

	// Version version of this service
	Version string `json:"version,omitempty"`

	// App The exposed app name, if no value is set, then metadata.name of the Service is used
	App string `json:"app,omitempty"`

	// Weight The weight among services with matching app field to determine how much traffic is load balanced
	// to this service.  If rollout is set, the weight become the target weight of the rollout.
	Weight *int `json:"weight,omitempty"`

	// Number of desired pods. This is a pointer to distinguish between explicit zero and not specified. Defaults to 1.
	Replicas *int `json:"replicas,omitempty" mapper:"alias=scale"`

	// The maximum number of pods that can be unavailable during the update.
	// Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// Absolute number is calculated from percentage by rounding down.
	// This can not be 0 if MaxSurge is 0.
	// Defaults to 25%.
	// Example: when this is set to 30%, the old ReplicaSet can be scaled down to 70% of desired pods
	// immediately when the rolling update starts. Once new pods are ready, old ReplicaSet
	// can be scaled down further, followed by scaling up the new ReplicaSet, ensuring
	// that the total number of pods available at all times during the update is at
	// least 70% of desired pods.
	// +optional
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty"`

	// The maximum number of pods that can be scheduled above the desired number of
	// pods.
	// Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// This can not be 0 if MaxUnavailable is 0.
	// Absolute number is calculated from percentage by rounding up.
	// Defaults to 25%.
	// Example: when this is set to 30%, the new ReplicaSet can be scaled up immediately when
	// the rolling update starts, such that the total number of old and new pods do not exceed
	// 130% of desired pods. Once old pods have been killed,
	// new ReplicaSet can be scaled up further, ensuring that total number of pods running
	// at any time during the update is at most 130% of desired pods.
	// +optional
	MaxSurge *intstr.IntOrString `json:"maxSurge,omitempty"`

	// Autoscale the replicas based on the amount of traffic received by this service
	Autoscale *AutoscaleConfig `json:"autoscale,omitempty"`

	// RolloutConfig controls how each service is allocated ComputedWeight
	RolloutConfig *RolloutConfig `json:"rollout,omitempty"`

	// Place one pod per node that matches the scheduling rules
	Global bool `json:"global,omitempty"`

	// Whether to disable Service mesh for Service. If true, no mesh sidecar will be deployed along with the Service
	ServiceMesh *bool `json:"serviceMesh,omitempty"`

	// Permissions to the Services. It will create corresponding ServiceAccounts, Roles and RoleBinding.
	Permissions []Permission `json:"permissions,omitempty" mapper:"permissions,alias=permission"`

	// GlobalPermissions to the Services. It will create corresponding ServiceAccounts, ClusterRoles and ClusterRoleBinding.
	GlobalPermissions []Permission `json:"globalPermissions,omitempty" mapper:"permissions,alias=globalPermission"`
}

type PodDNSConfigOption struct {
	Name  string  `json:"name,omitempty"`
	Value *string `json:"value,omitempty"`
}

// ContainerSecurityContext holds pod-level security attributes and common container constants. Optional: Defaults to empty. See type description for default values of each field.
type ContainerSecurityContext struct {
	// The UID to run the entrypoint of the container process. Defaults to user specified in image metadata if unspecified. May also be set in SecurityContext.
	// If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence for that container
	RunAsUser *int64 `json:"runAsUser,omitempty" mapper:"alias=user"`

	// The GID to run the entrypoint of the container process. Uses runtime default if unset. May also be set in SecurityContext.
	// If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence for that container.
	RunAsGroup *int64 `json:"runAsGroup,omitempty" mapper:"alias=user"`

	// Whether this container has a read-only root filesystem. Default is false.
	ReadOnlyRootFilesystem *bool `json:"readOnlyRootFilesystem,omitempty"`

	// Run container in privileged mode.
	// Processes in privileged containers are essentially equivalent to root on the host.
	// Defaults to false.
	// +optional
	Privileged *bool `json:"privileged,omitempty"`
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

	// ImageBuild specifies how to build this image
	ImageBuild *ImageBuildSpec `json:"build,omitempty"`

	// Entrypoint array. Not executed within a shell. The docker image's ENTRYPOINT is used if this is not provided.
	// Variable references $(VAR_NAME) are expanded using the container's environment. If a variable cannot be resolved, the reference in the input string will be unchanged.
	// The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not.
	// Cannot be updated. More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell
	Command []string `json:"command,omitempty" mapper:"shlex"`

	// Arguments to the entrypoint. The docker image's CMD is used if this is not provided.
	// Variable references $(VAR_NAME) are expanded using the container's environment.
	// If a variable cannot be resolved, the reference in the input string will be unchanged.
	// The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not.
	// Cannot be updated. More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell
	Args []string `json:"args,omitempty" mapper:"shlex,alias=arg"`

	// Container's working directory. If not specified, the container runtime's default will be used, which might be configured in the container image. Cannot be updated.
	WorkingDir string `json:"workingDir,omitempty"`

	// List of ports to expose from the container. Exposing a port here gives the system additional information about the network connections a container uses, but is primarily informational. Not specifying a port here DOES NOT prevent that port from being exposed.
	// Any port which is listening on the default "0.0.0.0" address inside a container will be accessible from the network. Cannot be updated.
	Ports []ContainerPort `json:"ports,omitempty" mapper:"ports,alias=port"`

	// List of environment variables to set in the container. Cannot be updated.
	Env []EnvVar `json:"env,omitempty" mapper:"env,envmap=sep==,alias=environment"`

	// CPU, in cores. (500m = .5 cores)
	CPUs *resource.Quantity `json:"cpus,omitempty" mapper:"quantity,alias=cpu"`

	// Memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)
	Memory *resource.Quantity `json:"memory,omitempty" mapper:"quantity,alias=mem"`

	// Secrets Mounts
	Secrets []DataMount `json:"secrets,omitempty" mapper:"secrets,envmap=sep=:,alias=secret"`

	// Configmaps Mounts
	Configs []DataMount `json:"configs,omitempty" mapper:"configs,envmap=sep=:,alias=config"`

	// Periodic probe of container liveness. Container will be restarted if the probe fails. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	LivenessProbe *v1.Probe `json:"livenessProbe,omitempty" mapper:"alias=liveness"`

	// Periodic probe of container service readiness. Container will be removed from service endpoints if the probe fails. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	ReadinessProbe *v1.Probe `json:"readinessProbe,omitempty" mapper:"alias=readiness"`

	// Image pull policy. One of Always, Never, IfNotPresent. Defaults to Always if :latest tag is specified, or IfNotPresent otherwise. Cannot be updated. More info: https://kubernetes.io/docs/concepts/containers/images#updating-images
	ImagePullPolicy v1.PullPolicy `json:"imagePullPolicy,omitempty" mapper:"enum=Always|IfNotPresent|Never,alias=pullPolicy"`

	// Whether this container should allocate a buffer for stdin in the container runtime. If this is not set, reads from stdin in the container will always result in EOF. Default is false.
	Stdin bool `json:"stdin,omitempty" mapper:"alias=stdinOpen"`

	// Whether the container runtime should close the stdin channel after it has been opened by a single attach. When stdin is true the stdin stream will remain open across multiple attach sessions.
	// If stdinOnce is set to true, stdin is opened on container start, is empty until the first client attaches to stdin, and then remains open and accepts data until the client disconnects, at which time stdin is closed and remains closed until the container is restarted. If this flag is false, a container processes that reads from stdin will never receive an EOF. Default is false
	StdinOnce bool `json:"stdinOnce,omitempty"`

	// Whether this container should allocate a TTY for itself, also requires 'stdin' to be true. Default is false.
	TTY bool `json:"tty,omitempty"`

	// Pod volumes to mount into the container's filesystem
	Volumes []Volume `json:"volumes,omitempty" mapper:"volumes,envmap=sep=:,alias=volume"`

	*ContainerSecurityContext
}

type DataMount struct {
	// The directory or file to mount the value to in the container
	Target string `json:"target,omitempty"`
	// The name of the ConfigMap or Secret to mount
	Name string `json:"name,omitempty"`
	// The key in the data of the ConfigMap or Secret to mount to a file.
	// If Key is set the Target must be a file.  If key is set the target must be a directory and will
	// contain one file per key from the Secret/ConfigMap data field.
	Key string `json:"key,omitempty"`
}

type VolumeTemplate struct {
	// Labels to be applied to the created PVC
	Labels map[string]string `json:"labels,omitempty"`
	// Annotations to be applied to the created PVC
	Annotations map[string]string `json:"annotations,omitempty"`

	// Name of the VolumeTemplate. A volume entry will use this name to refer to the created volume
	Name string

	// AccessModes contains the desired access modes the volume should have.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1
	// +optional
	AccessModes []v1.PersistentVolumeAccessMode `json:"accessModes,omitempty"`
	// Resources represents the minimum resources the volume should have.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources
	// +optional
	StorageRequest resource.Quantity `json:"storage,omitempty" mapper:"quantity"`
	// Name of the StorageClass required by the claim.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1
	StorageClassName string `json:"storageClassName,omitempty"`
	// volumeMode defines what type of volume is required by the claim.
	// Value of Filesystem is implied when not included in claim spec.
	// This is a beta feature.
	// +optional
	VolumeMode *v1.PersistentVolumeMode `json:"volumeMode,omitempty"`
}

type Volume struct {
	// Name is the name of the volume. If multiple Volumes in the same pod share the same name they
	// will be the same underlying storage. If persistent is set to true Name is required and will be
	// used to reference a PersistentVolumeClaim in the current namespace.
	//
	// If Name matches the name of a VolumeTemplate on this service then the VolumeTemplate will be used as the
	// source of the volume.
	Name string `json:"name,omitempty"`

	// That path within the container to mount the volume to
	Path string `json:"path,omitempty"`

	// That path on the host to mount into this container
	HostPath string `json:"hostpath,omitempty"`

	// The
	HostPathType *v1.HostPathType `json:"hostPathType,omitempty" protobuf:"bytes,2,opt,name=type"`

	// If Persistent is true then this volume refers to a PersistentVolumeClaim in this namespace. The
	// Name field is used to reference PersistentVolumeClaim.  If the Name of this Volume matches a VolumeTemplate
	// then Persistent is assumed to be true
	Persistent bool `json:"persistent,omitempty"`
}

type EnvVar struct {
	Name          string `json:"name,omitempty"`
	Value         string `json:"value,omitempty"`
	SecretName    string `json:"secretName,omitempty"`
	ConfigMapName string `json:"configMapName,omitempty"`
	Key           string `json:"key,omitempty"`
}

type DNS struct {
	// Set DNS policy for the pod. Defaults to "ClusterFirst". Valid values are 'ClusterFirstWithHostNet', 'ClusterFirst', 'Default' or 'None'.
	// DNS parameters given in DNSConfig will be merged with the policy selected with DNSPolicy.
	// To have DNS options set along with hostNetwork, you have to specify DNS policy explicitly to 'ClusterFirstWithHostNet'.
	Policy v1.DNSPolicy `json:"policy,omitempty" mapper:"enum=ClusterFirst|ClusterFirstWithHostNet|Default|None|host=Default"`

	// A list of DNS name server IP addresses. This will be appended to the base nameservers generated from DNSPolicy. Duplicated nameservers will be removed.
	Nameservers []string `json:"nameservers,omitempty"`

	// A list of DNS search domains for host-name lookup. This will be appended to the base search paths generated from DNSPolicy. Duplicated search paths will be removed.
	Searches []string `json:"searches,omitempty"`

	// A list of DNS resolver options. This will be merged with the base options generated from DNSPolicy.
	// Duplicated entries will be removed. Resolution options given in Options will override those that appear in the base DNSPolicy.
	Options []PodDNSConfigOption `json:"options,omitempty" mapper:"dnsOptions"`
}

type PodConfig struct {
	// List of containers belonging to the pod. Containers cannot currently be added or removed. There must be at least one container in a Pod. Cannot be updated.
	Sidecars []NamedContainer `json:"containers,omitempty"`

	// Specifies the hostname of the Pod If not specified, the pod's hostname will be set to a system-defined value.
	Hostname string `json:"hostname,omitempty"`

	// HostAliases is an optional list of hosts and IPs that will be injected into the pod's hosts file if specified. This is only valid for non-hostNetwork pods.
	HostAliases []v1.HostAlias `json:"hostAliases,omitempty" mapper:"hosts,envmap=sep==,alias=hosts"`

	// Host networking requested for this pod. Use the host's network namespace. If this option is set, the ports that will be used must be specified. Default to false.
	HostNetwork bool `json:"hostNetwork,omitempty" mapper:"hostNetwork"`

	// Image pull secret
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty" mapper:"alias=pullSecrets"`

	// Volumes to create per replica
	VolumeTemplates []VolumeTemplate `json:"volumeTemplates,omitempty"`

	// DNS settings for this Pod
	DNS *DNS `json:"dns,omitempty"`

	*v1.Affinity `json:",inline"`

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
	Name string `json:"name,omitempty"`
	// Expose will make the port available outside the cluster. All http/https ports will be set to true by default
	// if Expose is nil.  All other protocols are set to false by default
	Expose     *bool    `json:"expose,omitempty"`
	Protocol   Protocol `json:"protocol,omitempty"`
	Port       int32    `json:"port"`
	TargetPort int32    `json:"targetPort,omitempty"`
	HostPort   bool     `json:"hostport,omitempty"`
}

func (in ContainerPort) IsHTTP() bool {
	return in.Protocol == "" || in.Protocol == ProtocolHTTP || in.Protocol == ProtocolHTTP2
}

func (in ContainerPort) IsExposed() bool {
	if in.Expose != nil {
		return *in.Expose
	}
	return in.IsHTTP()
}

type ServiceStatus struct {
	// DeploymentReady for ready status on deployment
	DeploymentReady bool `json:"deploymentReady,omitempty"`

	// ScaleStatus for the Service
	ScaleStatus *ScaleStatus `json:"scaleStatus,omitempty"`

	// ComputedReplicas is calculated from autoscaling component to make sure pod has the desired load
	ComputedReplicas *int `json:"computedReplicas,omitempty"`

	// ComputedWeight is the weight calculated from the rollout revision
	ComputedWeight *int `json:"computedWeight,omitempty"`

	// ContainerImages are populated from builds to override the images of this service
	ContainerImages map[string]BuiltImage `json:"containerImages,omitempty"`

	// Represents the latest available observations of a deployment's current state.
	Conditions []genericcondition.GenericCondition `json:"conditions,omitempty"`

	// The Endpoints to access this version directly
	Endpoints []string `json:"endpoints,omitempty" column:"name=Endpoint,type=string,jsonpath=.status.endpoints[0]"`

	// The Endpoints to access this service as part of an app
	AppEndpoints []string `json:"appEndpoints,omitempty"`

	// log token to access build log
	BuildLogToken string `json:"buildLogToken,omitempty"`

	// Associated git commit name
	GitCommitName string `json:"gitCommitName,omitempty"`
}

type BuiltImage struct {
	ImageName  string `json:"imageName,omitempty"`
	PullSecret string `json:"pullSecret,omitempty"`
}

type ScaleStatus struct {
	// Total number of unavailable pods targeted by this deployment. This is the total number of pods that are still required for the deployment to have 100% available capacity.
	// They may either be pods that are running but not yet available or pods that still have not been created.
	Unavailable int `json:"unavailable,omitempty"`

	// Total number of available pods (ready for at least minReadySeconds) targeted by this deployment.
	Available int `json:"available,omitempty"`
}
