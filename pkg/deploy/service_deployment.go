package deploy

import (
	"strings"

	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func isDeployment(service *v1beta1.ServiceUnversionedSpec, usedTemplates map[string]*v1beta1.Volume) bool {
	if service.UpdateStrategy == "on-delete" || service.DeploymentStrategy == "ordered" {
		return false
	}

	if len(usedTemplates) > 0 {
		return false
	}

	return true
}

func mergeLabels(base, overlay map[string]string) map[string]string {
	result := map[string]string{}
	for k, v := range base {
		result[k] = v
	}

	for k, v := range overlay {
		if strings.HasPrefix(k, "rio.cattle.io") {
			continue
		}
		result[k] = v
	}

	return result
}

func deployment(objects []runtime.Object, labels map[string]string, depName, namespace string, service *v1beta1.ServiceUnversionedSpec, podTemplateSpec v1.PodTemplateSpec) []runtime.Object {
	scaleParams := parseScaleParams(service)

	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        depName,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: map[string]string{},
		},
		Spec: appsv1.DeploymentSpec{
			Paused:   false,
			Replicas: &scaleParams.Scale,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: podTemplateSpec,
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
		},
	}

	if scaleParams.Scale > 0 && scaleParams.BatchSize > 0 {
		dep.Spec.Strategy.RollingUpdate = &appsv1.RollingUpdateDeployment{
			MaxSurge:       scaleParams.MaxSurge,
			MaxUnavailable: scaleParams.MaxUnavailable,
		}
	}

	return append(objects, dep)
}
