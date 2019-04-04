package podcontrollers

import (
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/objectset"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func isDeployment(service riov1.ServiceSpec) bool {
	if service.UpdateStrategy == "on-delete" || service.DeploymentStrategy == "ordered" {
		return false
	}

	if len(service.VolumeTemplates) > 0 {
		return false
	}

	return true
}

func deployment(service *riov1.Service, cp *controllerParams, os *objectset.ObjectSet) {
	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        service.Name,
			Namespace:   service.Namespace,
			Labels:      cp.Labels,
			Annotations: map[string]string{},
		},
		Spec: appsv1.DeploymentSpec{
			Paused:   false,
			Replicas: &cp.Scale.Scale,
			Selector: &metav1.LabelSelector{
				MatchLabels: cp.SelectorLabels,
			},
			Template: cp.PodTemplateSpec,
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
		},
	}

	if cp.Scale.Scale > 0 && cp.Scale.BatchSize > 0 {
		dep.Spec.Strategy.RollingUpdate = &appsv1.RollingUpdateDeployment{
			MaxSurge:       cp.Scale.MaxSurge,
			MaxUnavailable: cp.Scale.MaxUnavailable,
		}
	}

	os.Add(dep)
}
