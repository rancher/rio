package deploy

import (
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func statefulset(objects []runtime.Object, labels map[string]string, depName, namespace string, service *v1beta1.ServiceUnversionedSpec, usedTemplates map[string]*v1beta1.Volume, podTemplateSpec v1.PodTemplateSpec) ([]runtime.Object, error) {
	scaleParams := parseScaleParams(service)

	statefulSet := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        depName,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: map[string]string{},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    &scaleParams.Scale,
			ServiceName: depName,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: podTemplateSpec,
		},
	}

	if service.UpdateStrategy == "on-delete" {
		statefulSet.Spec.UpdateStrategy.Type = appsv1.OnDeleteStatefulSetStrategyType
	} else {
		zero := int32(0)
		statefulSet.Spec.UpdateStrategy.Type = appsv1.RollingUpdateStatefulSetStrategyType
		statefulSet.Spec.UpdateStrategy.RollingUpdate = &appsv1.RollingUpdateStatefulSetStrategy{
			Partition: &zero,
		}
	}

	if service.DeploymentStrategy == "ordered" {
		statefulSet.Spec.PodManagementPolicy = appsv1.OrderedReadyPodManagement
	} else {
		statefulSet.Spec.PodManagementPolicy = appsv1.RollingUpdateStatefulSetStrategyType
	}

	for _, volumeTemplate := range usedTemplates {
		labels := map[string]string{
			"rio.cattle.io/namespace":      namespace,
			"rio.cattle.io/volumetemplate": volumeTemplate.Name,
		}

		pvc, err := volumeToPVC(namespace, labels, *volumeTemplate)
		if err != nil {
			return nil, err
		}

		statefulSet.Spec.VolumeClaimTemplates = append(statefulSet.Spec.VolumeClaimTemplates, *pvc)
	}

	return append(objects, statefulSet), nil
}
