package flat

import (
	"testing"

	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/conversion"
)

var (
	five = int64(5)
	six  = int32(5)
	t    = true
	blah = "blah"

	FlatService = ServiceSpec{
		Global:             true,
		BatchSize:          1,
		DeploymentStrategy: "foo",
		UpdateOrder:        "foo",
		AutoScale: &AutoScale{
			MinScale:    1,
			MaxScale:    2,
			Concurrency: 1,
		},
		Permissions: []string{
			"role=foo",
			"role=",
			"read foo",
			"write /node/status bar",
			"rule=get,put group,group2/a,c/status,d bar,baz",
		},
		GlobalPermissions: []string{
			"role=foo",
			"role=",
			"read foo",
			"write /node/status bar",
			"rule=get,put group,group2/a,c/status,d bar,baz",
		},
		PodSpec: PodSpec{
			RestartPolicy:                 "foo",
			TerminationGracePeriodSeconds: &five,
			ActiveDeadlineSeconds:         &five,
			DNSPolicy:                     "foo",
			NodeSelector: map[string]string{
				"foo": "bar",
			},
			ServiceAccountName:           "foo",
			DeprecatedServiceAccount:     "foo",
			AutomountServiceAccountToken: &t,
			NodeName:                     "foo",
			HostNetwork:                  true,
			HostPID:                      true,
			HostIPC:                      true,
			ShareProcessNamespace:        &t,
			Hostname:                     "foo",
			Subdomain:                    "foo",
			SchedulerName:                "foo",
			PriorityClassName:            "foo",
			Priority:                     &six,
			RuntimeClassName:             &blah,
			EnableServiceLinks:           &t,
		},
		Container: Container{
			Name:       "foo",
			Image:      "foo",
			Command:    []string{"foo"},
			Args:       []string{"foo"},
			WorkingDir: "foo",
			//Ports []ContainerPort `json:"ports,omitempty" patchStrategy:"merge" patchMergeKey:"containerPort" protobuf:"bytes,6,rep,name=ports"`
			//EnvFrom []EnvFromSource `json:"envFrom,omitempty" protobuf:"bytes,19,rep,name=envFrom"`
			//Env []EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,7,rep,name=env"`
			//Resources ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
			//VolumeMounts []VolumeMount `json:"volumeMounts,omitempty" patchStrategy:"merge" patchMergeKey:"mountPath" protobuf:"bytes,9,rep,name=volumeMounts"`
			//VolumeDevices []VolumeDevice `json:"volumeDevices,omitempty" patchStrategy:"merge" patchMergeKey:"devicePath" protobuf:"bytes,21,rep,name=volumeDevices"`
			//LivenessProbe *Probe `json:"livenessProbe,omitempty" protobuf:"bytes,10,opt,name=livenessProbe"`
			//ReadinessProbe *Probe `json:"readinessProbe,omitempty" protobuf:"bytes,11,opt,name=readinessProbe"`
			//Lifecycle                *Lifecycle               `json:"lifecycle,omitempty" protobuf:"bytes,12,opt,name=lifecycle"`
			TerminationMessagePath:   "foo",
			TerminationMessagePolicy: "foo",
			ImagePullPolicy:          "foo",
			//SecurityContext          *SecurityContext `json:"securityContext,omitempty" protobuf:"bytes,15,opt,name=securityContext"`
			Stdin:     true,
			StdinOnce: true,
			TTY:       true,
		},
	}
	ExpandService = v1.ServiceSpec{
		Global:             true,
		BatchSize:          1,
		DeploymentStrategy: "foo",
		UpdateOrder:        "foo",
		AutoScale: &v1.AutoscaleConfig{
			MinScale:    1,
			MaxScale:    2,
			Concurrency: 1,
		},
		Roles: []string{"foo"},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Verbs: []string{
					"get",
					"list",
					"watch",
				},
				Resources: []string{"foo"},
			},
			{
				APIGroups: []string{""},
				Verbs: []string{
					"create",
					"delete",
					"get",
					"list",
					"patch",
					"update",
					"watch",
				},
				Resources:     []string{"node/status"},
				ResourceNames: []string{"bar"},
			},
			{
				APIGroups: []string{"group", "group2"},
				Verbs: []string{
					"get",
					"put",
				},
				Resources: []string{
					"a",
					"c/status",
					"d",
				},
				ResourceNames: []string{
					"bar",
					"baz",
				},
			},
		},
		ClusterRoles: []string{"foo"},
		ClusterRules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Verbs: []string{
					"get",
					"list",
					"watch",
				},
				Resources: []string{"foo"},
			},
			{
				APIGroups: []string{""},
				Verbs: []string{
					"create",
					"delete",
					"get",
					"list",
					"patch",
					"update",
					"watch",
				},
				Resources:     []string{"node/status"},
				ResourceNames: []string{"bar"},
			},
			{
				APIGroups: []string{"group", "group2"},
				Verbs: []string{
					"get",
					"put",
				},
				Resources: []string{
					"a",
					"c/status",
					"d",
				},
				ResourceNames: []string{
					"bar",
					"baz",
				},
			},
		},
		PodSpec: corev1.PodSpec{
			RestartPolicy:                 "foo",
			TerminationGracePeriodSeconds: &five,
			ActiveDeadlineSeconds:         &five,
			DNSPolicy:                     "foo",
			NodeSelector: map[string]string{
				"foo": "bar",
			},
			ServiceAccountName:           "foo",
			DeprecatedServiceAccount:     "foo",
			AutomountServiceAccountToken: &t,
			NodeName:                     "foo",
			HostNetwork:                  true,
			HostPID:                      true,
			HostIPC:                      true,
			ShareProcessNamespace:        &t,
			Hostname:                     "foo",
			Subdomain:                    "foo",
			SchedulerName:                "foo",
			PriorityClassName:            "foo",
			Priority:                     &six,
			RuntimeClassName:             &blah,
			EnableServiceLinks:           &t,
			Containers: []corev1.Container{
				{
					Name:       "foo",
					Image:      "foo",
					Command:    []string{"foo"},
					Args:       []string{"foo"},
					WorkingDir: "foo",
					Ports: []corev1.ContainerPort{
						{
							Name:          "portname",
							HostPort:      1024,
							ContainerPort: 11024,
							Protocol:      "UDP",
							HostIP:        "1.1.1.1",
						},
						{
							Name:          "portname2",
							HostPort:      1025,
							ContainerPort: 11025,
							Protocol:      "TCP",
							HostIP:        "1.1.1.1",
						},
					},
					//Ports []ContainerPort `json:"ports,omitempty" patchStrategy:"merge" patchMergeKey:"containerPort" protobuf:"bytes,6,rep,name=ports"`
					//EnvFrom []EnvFromSource `json:"envFrom,omitempty" protobuf:"bytes,19,rep,name=envFrom"`
					//Env []EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,7,rep,name=env"`
					//Resources ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
					//VolumeMounts []VolumeMount `json:"volumeMounts,omitempty" patchStrategy:"merge" patchMergeKey:"mountPath" protobuf:"bytes,9,rep,name=volumeMounts"`
					//VolumeDevices []VolumeDevice `json:"volumeDevices,omitempty" patchStrategy:"merge" patchMergeKey:"devicePath" protobuf:"bytes,21,rep,name=volumeDevices"`
					//LivenessProbe *Probe `json:"livenessProbe,omitempty" protobuf:"bytes,10,opt,name=livenessProbe"`
					//ReadinessProbe *Probe `json:"readinessProbe,omitempty" protobuf:"bytes,11,opt,name=readinessProbe"`
					//Lifecycle                *Lifecycle               `json:"lifecycle,omitempty" protobuf:"bytes,12,opt,name=lifecycle"`
					TerminationMessagePath:   "foo",
					TerminationMessagePolicy: "foo",
					ImagePullPolicy:          "foo",
					//SecurityContext          *SecurityContext `json:"securityContext,omitempty" protobuf:"bytes,15,opt,name=securityContext"`
					Stdin:     true,
					StdinOnce: true,
					TTY:       true,
				},
			},
		},
	}
)

func TestService(t *testing.T) {
	serviceSpec := &v1.ServiceSpec{}
	err := converter.Convert(&FlatService, serviceSpec, conversion.SourceToDest, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !equality.Semantic.DeepEqual(&ExpandService, serviceSpec) {
		t.Fatalf("Do not match %+v %+v", &ExpandService, serviceSpec)
	}
}
