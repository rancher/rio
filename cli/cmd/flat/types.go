package flat

import (
	"reflect"

	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/pretty/stringers"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/util/runtime"
)

var (
	expandFlags = conversion.SourceToDest | conversion.AllowDifferentFieldTypeNames
	converter   = conversion.NewConverter(func(t reflect.Type) string {
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		return t.Name()
	})
)

func init() {
	runtime.Must(converter.RegisterConversionFunc(ExpandServiceSpec))
	runtime.Must(converter.RegisterConversionFunc(ExpandPolicyRule))
}

type ServiceSpec struct {
	Scale              int
	BatchSize          int
	UpdateOrder        string
	UpdateStrategy     string
	Global             bool
	DeploymentStrategy string

	*AutoScale

	Permissions       []string
	GlobalPermissions []string

	PodSpec
	Container
}

type Container struct {
	Name       string   `json:"name" protobuf:"bytes,1,opt,name=name"`
	Image      string   `json:"image,omitempty" protobuf:"bytes,2,opt,name=image"`
	Command    []string `json:"command,omitempty" protobuf:"bytes,3,rep,name=command"`
	Args       []string `json:"args,omitempty" protobuf:"bytes,4,rep,name=args"`
	WorkingDir string   `json:"workingDir,omitempty" protobuf:"bytes,5,opt,name=workingDir"`
	//Ports []ContainerPort `json:"ports,omitempty" patchStrategy:"merge" patchMergeKey:"containerPort" protobuf:"bytes,6,rep,name=ports"`
	//EnvFrom []EnvFromSource `json:"envFrom,omitempty" protobuf:"bytes,19,rep,name=envFrom"`
	//Env []EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,7,rep,name=env"`
	//Resources ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
	//VolumeMounts []VolumeMount `json:"volumeMounts,omitempty" patchStrategy:"merge" patchMergeKey:"mountPath" protobuf:"bytes,9,rep,name=volumeMounts"`
	//VolumeDevices []VolumeDevice `json:"volumeDevices,omitempty" patchStrategy:"merge" patchMergeKey:"devicePath" protobuf:"bytes,21,rep,name=volumeDevices"`
	//LivenessProbe *Probe `json:"livenessProbe,omitempty" protobuf:"bytes,10,opt,name=livenessProbe"`
	//ReadinessProbe *Probe `json:"readinessProbe,omitempty" protobuf:"bytes,11,opt,name=readinessProbe"`
	//Lifecycle                *Lifecycle               `json:"lifecycle,omitempty" protobuf:"bytes,12,opt,name=lifecycle"`
	TerminationMessagePath   string `json:"terminationMessagePath,omitempty" protobuf:"bytes,13,opt,name=terminationMessagePath"`
	TerminationMessagePolicy string `json:"terminationMessagePolicy,omitempty" protobuf:"bytes,20,opt,name=terminationMessagePolicy,casttype=TerminationMessagePolicy"`
	ImagePullPolicy          string `json:"imagePullPolicy,omitempty" protobuf:"bytes,14,opt,name=imagePullPolicy,casttype=PullPolicy"`
	//SecurityContext          *SecurityContext `json:"securityContext,omitempty" protobuf:"bytes,15,opt,name=securityContext"`
	Stdin     bool `json:"stdin,omitempty" protobuf:"varint,16,opt,name=stdin"`
	StdinOnce bool `json:"stdinOnce,omitempty" protobuf:"varint,17,opt,name=stdinOnce"`
	TTY       bool `json:"tty,omitempty" protobuf:"varint,18,opt,name=tty"`
}

type PodSpec struct {
	//Volumes                       []Volume               `json:"volumes,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name" protobuf:"bytes,1,rep,name=volumes"`
	//InitContainers                []Container            `json:"initContainers,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,20,rep,name=initContainers"`
	//Containers                    []Container            `json:"containers" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,2,rep,name=containers"`
	RestartPolicy                 string
	TerminationGracePeriodSeconds *int64
	ActiveDeadlineSeconds         *int64
	DNSPolicy                     string
	NodeSelector                  map[string]string
	ServiceAccountName            string
	DeprecatedServiceAccount      string
	AutomountServiceAccountToken  *bool
	NodeName                      string
	HostNetwork                   bool
	HostPID                       bool
	HostIPC                       bool
	ShareProcessNamespace         *bool
	//SecurityContext               *PodSecurityContext
	//ImagePullSecrets              []LocalObjectReference
	Hostname  string
	Subdomain string
	//Affinity           *Affinity
	SchedulerName string
	//Tolerations        []Toleration
	//HostAliases        []HostAlias
	PriorityClassName string
	Priority          *int32
	//DNSConfig          *PodDNSConfig
	//ReadinessGates     []PodReadinessGate
	RuntimeClassName   *string
	EnableServiceLinks *bool
}

type AutoScale struct {
	Concurrency int
	MinScale    int
	MaxScale    int
}

func defaultExpand(a, b interface{}, scope conversion.Scope) error {
	return scope.DefaultConvert(a, b, expandFlags)
}

func ExpandPolicyRule(rules *[]string, policyRules *[]rbacv1.PolicyRule, scope conversion.Scope) error {
	return nil
}

func ExpandServiceSpec(flat *ServiceSpec, expanded *v1.ServiceSpec, scope conversion.Scope) error {
	var (
		err error
	)

	if err := scope.DefaultConvert(flat, expanded, expandFlags|conversion.IgnoreMissingFields); err != nil {
		return err
	}

	expanded.Roles = stringers.ParseRoles(flat.Permissions...)
	expanded.ClusterRoles = stringers.ParseRoles(flat.GlobalPermissions...)

	expanded.Rules, err = stringers.ParsePolicyRules(flat.Permissions...)
	if err != nil {
		return err
	}

	expanded.ClusterRules, err = stringers.ParsePolicyRules(flat.GlobalPermissions...)
	if err != nil {
		return err
	}

	return nil
}
