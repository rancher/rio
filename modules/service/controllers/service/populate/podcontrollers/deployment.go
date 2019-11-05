package podcontrollers

import (
	"fmt"
	"sort"

	"k8s.io/apimachinery/pkg/api/resource"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/wrangler/pkg/objectset"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func statefulset(service *riov1.Service, cp *controllerParams, os *objectset.ObjectSet) {
	appName, version := services.AppAndVersion(service)

	ss := constructors.NewStatefulSet(service.Namespace, service.Name, appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      cp.Labels,
			Annotations: cp.Annotations,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: nil,
			Selector: &metav1.LabelSelector{
				MatchLabels: cp.SelectorLabels,
			},
			Template:             cp.PodTemplateSpec,
			VolumeClaimTemplates: volumeClaimTemplates(cp.VolumeTemplates),
			ServiceName:          fmt.Sprintf("%s-%s", appName, version),
			PodManagementPolicy:  appsv1.ParallelPodManagement,
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
		},
	})

	os.Add(ss)
}

func deployment(service *riov1.Service, cp *controllerParams, os *objectset.ObjectSet) {
	dep := constructors.NewDeployment(service.Namespace, service.Name, appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      cp.Labels,
			Annotations: cp.Annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: cp.Scale.Scale,
			Selector: &metav1.LabelSelector{
				MatchLabels: cp.SelectorLabels,
			},
			Template: cp.PodTemplateSpec,
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: cp.Scale.MaxUnavailable,
					MaxSurge:       cp.Scale.MaxSurge,
				},
			},
		},
	})

	os.Add(dep)
}

func daemonset(service *riov1.Service, cp *controllerParams, os *objectset.ObjectSet) {
	ds := constructors.NewDaemonset(service.Namespace, service.Name, appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      cp.Labels,
			Annotations: cp.Annotations,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: cp.SelectorLabels,
			},
			Template: cp.PodTemplateSpec,
			UpdateStrategy: appsv1.DaemonSetUpdateStrategy{
				Type: appsv1.RollingUpdateDaemonSetStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDaemonSet{
					MaxUnavailable: cp.Scale.MaxUnavailable,
				},
			},
		},
	})

	os.Add(ds)
}

func volumeClaimTemplates(templates map[string]riov1.VolumeTemplate) (result []v1.PersistentVolumeClaim) {
	var names []string
	for name := range templates {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		template := templates[name]
		q := resource.NewQuantity(template.StorageRequest, resource.BinarySI)
		result = append(result, v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "vol-" + name,
				Labels:      template.Labels,
				Annotations: template.Annotations,
			},
			Spec: v1.PersistentVolumeClaimSpec{
				AccessModes: template.AccessModes,
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceStorage: *q,
					},
				},
				StorageClassName: &template.StorageClassName,
				VolumeMode:       template.VolumeMode,
			},
		})
	}

	return
}
