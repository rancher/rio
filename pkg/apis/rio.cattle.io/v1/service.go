package v1

import (
	"github.com/rancher/wrangler/pkg/genericcondition"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	Version       string `json:"version,omitempty"`
	Weight        int    `json:"weight,omitempty"`
	Promote       bool   `json:"promote,omitempty"`
	App           string `json:"app,omitempty"`
	ParentService string `json:"parentService,omitempty"`
}

type AutoscaleConfig struct {
	Concurrency int `json:"concurrency,omitempty" norman:"default=10"`
	MinScale    int `json:"minScale,omitempty" norman:"default=1"`
	MaxScale    int `json:"maxScale,omitempty" norman:"default=30"`
}

type ServiceSpec struct {
	Scale              int              `json:"scale,omitempty"`
	BatchSize          int              `json:"batchSize,omitempty"`
	UpdateOrder        string           `json:"updateOrder,omitempty" norman:"type=enum,options=start-first|stop-first"`
	UpdateStrategy     string           `json:"updateStrategy,omitempty" norman:"type=enum,options=rolling|on-delete,default=rolling"`
	Global             bool             `json:"global,omitempty"`
	DeploymentStrategy string           `json:"deploymentStrategy,omitempty" norman:"type=enum,options=parallel|ordered,default=parallel"`
	AutoScale          *AutoscaleConfig `json:"autoScale,omitempty"`
	DisableServiceMesh bool             `json:"disableServiceMesh,omitempty"`

	Roles        []string            `json:"roles,omitempty"`
	Rules        []rbacv1.PolicyRule `json:"rules,omitempty"`
	ClusterRoles []string            `json:"clusterRoles,omitempty"`
	ClusterRules []rbacv1.PolicyRule `json:"clusterRules,omitempty"`

	Revision ServiceRevision `json:"revision,omitempty"`

	ImageBuilds     []ContainerImageBuild      `json:"imageBuilds,omitempty"`
	PodSpec         v1.PodSpec                 `json:"podSpec,omitempty"`
	VolumeTemplates []v1.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty"`
	Ports           []ServicePort              `json:"ports,omitempty"`
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

type ServicePort struct {
	Name       string             `json:"name,omitempty"`
	Publish    bool               `json:"publish,omitempty"`
	IP         string             `json:"ip,omitempty"`
	Protocol   Protocol           `json:"protocol,omitempty" protobuf:"bytes,2,opt,name=protocol,casttype=Protocol"`
	Port       int32              `json:"port" protobuf:"varint,3,opt,name=port"`
	TargetPort intstr.IntOrString `json:"targetPort,omitempty" protobuf:"bytes,4,opt,name=targetPort"`
	NodePort   int32              `json:"nodePort,omitempty" protobuf:"varint,5,opt,name=nodePort"`
}

type ServiceStatus struct {
	DeploymentStatus  *appsv1.DeploymentStatus            `json:"deploymentStatus,omitempty"`
	DaemonSetStatus   *appsv1.DaemonSetStatus             `json:"daemonSetStatus,omitempty"`
	StatefulSetStatus *appsv1.StatefulSetStatus           `json:"statefulSetStatus,omitempty"`
	ScaleStatus       *ScaleStatus                        `json:"scaleStatus,omitempty"`
	ContainerImages   map[string]string                   `json:"containerImages,omitempty"`
	Conditions        []genericcondition.GenericCondition `json:"conditions,omitempty"`
	Endpoints         []Endpoint                          `json:"endpoints,omitempty"`
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

type ContainerImageBuild struct {
	ContainerName string `json:"containerName,omitempty"`

	ImageBuild
}

type ImageBuild struct {
	URL        string `json:"url,omitempty"`
	Tag        string `json:"tag,omitempty"`
	Commit     string `json:"commit,omitempty"`
	Branch     string `json:"branch,omitempty"`
	DockerFile string `json:"dockerFile,omitempty"`
	Template   string `json:"template,omitempty"`
	Secret     string `json:"secret,omitempty"`
	Hook       bool   `json:"hook,omitempty"`
}
