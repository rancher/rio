package populate

import (
	"testing"

	"github.com/rancher/rio/modules/service/controllers/service/populate/rbac"
	"github.com/rancher/rio/modules/test"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
	"gotest.tools/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestService(t *testing.T) {
	os := objectset.NewObjectSet()

	systemNs := "rio-system-fake"
	input := riov1.NewService("default", "test", riov1.Service{
		Spec: riov1.ServiceSpec{
			DisableServiceMesh: true,
			AutoscaleConfig: riov1.AutoscaleConfig{
				MaxScale:    &[]int{10}[0],
				MinScale:    &[]int{1}[0],
				Concurrency: &[]int{10}[0],
			},
			ServiceRevision: riov1.ServiceRevision{
				App:     "foo",
				Version: "v0",
			},
			GlobalPermissions: []riov1.Permission{
				{
					APIGroup: "test",
					Resource: "test-foo",
					Verbs:    []string{"GET", "CREATE"},
				},
			},
			Permissions: []riov1.Permission{
				{
					APIGroup: "test",
					Resource: "test-foo",
					Verbs:    []string{"GET", "CREATE"},
				},
			},
			ServiceScale: riov1.ServiceScale{
				Scale: &[]int{1}[0],
			},
			PodConfig: riov1.PodConfig{
				Container: riov1.Container{
					Image: "test",
					Args: []string{
						"echo",
						"hello world",
					},
					Env: []riov1.EnvVar{
						{
							Name:  "FOO",
							Value: "BAR",
						},
						{
							Name:       "SECRET",
							SecretName: "secret-foo",
							Key:        "secret-key",
						},
						{
							Name:          "CONFIG",
							ConfigMapName: "config-foo",
							Key:           "config-key",
						},
					},
					Secrets: []riov1.DataMount{
						{
							Name:      "secret-mount",
							Directory: "/path/secret-data",
							Key:       "password",
							File:      "password",
						},
					},
					Configs: []riov1.DataMount{
						{
							Name:      "config-mount",
							Directory: "/path/config-data",
							Key:       "config",
							File:      "config",
						},
					},
					Ports: []riov1.ContainerPort{
						{
							Port:       80,
							TargetPort: 80,
						},
					},
				},
			},
		},
	})

	maxUnavailable := intstr.FromInt(0)
	maxSurge := intstr.FromInt(1)
	deployment := constructors.NewDeployment("default", "test", appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
			Labels: map[string]string{
				"app":     "foo",
				"version": "v0",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: &maxUnavailable,
					MaxSurge:       &maxSurge,
				},
			},
			Replicas: &[]int32{1}[0],
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":     "foo",
					"version": "v0",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
					Labels: map[string]string{
						"app":     "foo",
						"version": "v0",
					},
				},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							Name: "secret-secret-mount",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: "secret-mount",
									Optional:   &[]bool{true}[0],
								},
							},
						},
						{
							Name: "config-config-mount",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "config-mount",
									},
								},
							},
						},
					},
					ServiceAccountName: "test",
					EnableServiceLinks: &[]bool{false}[0],
					Containers: []v1.Container{
						{
							Name:  "test",
							Image: "test",
							Env: []v1.EnvVar{
								{
									Name:  "FOO",
									Value: "BAR",
								},
								{
									Name: "SECRET",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											Key: "secret-key",
											LocalObjectReference: v1.LocalObjectReference{
												Name: "secret-foo",
											},
										},
									},
								},
								{
									Name: "CONFIG",
									ValueFrom: &v1.EnvVarSource{
										ConfigMapKeyRef: &v1.ConfigMapKeySelector{
											Key: "config-key",
											LocalObjectReference: v1.LocalObjectReference{
												Name: "config-foo",
											},
										},
									},
								},
							},
							Args: []string{
								"echo",
								"hello world",
							},
							Ports: []v1.ContainerPort{
								{
									Protocol:      v1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "config-config-mount",
									MountPath: "/path/config-data/config",
									SubPath:   "config",
								},
								{
									Name:      "secret-secret-mount",
									MountPath: "/path/secret-data/password",
									SubPath:   "password",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			},
		},
	})

	sa := constructors.NewServiceAccount("default", "test", v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
			Labels: map[string]string{
				"app":     "foo",
				"version": "v0",
			},
		},
	})
	clusterRole := rbac.NewClusterRole("rio-default-test", map[string]string{
		"app":     "foo",
		"version": "v0",
	})
	clusterRole.Rules = []rbacv1.PolicyRule{
		{
			Verbs: []string{
				"GET",
				"CREATE",
			},
			APIGroups: []string{
				"test",
			},
			Resources: []string{
				"test-foo",
			},
		},
	}
	role := rbac.NewRole("default", "rio-test", map[string]string{
		"app":     "foo",
		"version": "v0",
	})
	role.Rules = []rbacv1.PolicyRule{
		{
			Verbs: []string{
				"GET",
				"CREATE",
			},
			APIGroups: []string{
				"test",
			},
			Resources: []string{
				"test-foo",
			},
		},
	}
	clusterRoleBinding := rbac.NewClusterBinding("rio-default-test-rio-default-test", map[string]string{
		"app":     "foo",
		"version": "v0",
	})
	clusterRoleBinding.RoleRef = rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     "rio-default-test",
	}
	clusterRoleBinding.Subjects = []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      "test",
			Namespace: "default",
		},
	}
	roleBinding := rbac.NewBinding("default", "rio-test-rio-test", map[string]string{
		"app":     "foo",
		"version": "v0",
	})
	roleBinding.RoleRef = rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "Role",
		Name:     "rio-test",
	}
	roleBinding.Subjects = []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      "test",
			Namespace: "default",
		},
	}

	if err := Service(input, systemNs, os); err != nil {
		t.Fatal(err)
	}

	o := test.GetObject(t, deployment, os)
	deploy := o.(*appsv1.Deployment)
	deploy.Spec.Template.Spec.Containers[0].Resources = v1.ResourceRequirements{}
	assert.DeepEqual(t, deployment, deploy)

	test.AssertObjects(t, sa, os)
	test.AssertObjects(t, role, os)
	test.AssertObjects(t, roleBinding, os)
	test.AssertObjects(t, clusterRole, os)
	test.AssertObjects(t, clusterRoleBinding, os)
}
