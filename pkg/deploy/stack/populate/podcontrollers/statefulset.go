package podcontrollers

import (
	"github.com/rancher/rio/pkg/deploy/stack/input"
	"github.com/rancher/rio/pkg/deploy/stack/output"
	"github.com/rancher/rio/pkg/deploy/stack/populate/volume"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	appsv1 "k8s.io/api/apps/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const useTemplatesLabel = "rio.cattle.io/use-templates"

func statefulSet(stack *input.Stack, service *v1beta1.Service, cp *controllerParams, usedTemplates map[string]*v1beta1.Volume, output *output.Deployment) error {
	statefulSet := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1beta2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        service.Name,
			Namespace:   stack.Namespace,
			Labels:      cp.Labels,
			Annotations: map[string]string{},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    &cp.Scale.Scale,
			ServiceName: service.Name,
			Selector: &metav1.LabelSelector{
				MatchLabels: cp.SelectorLabels,
			},
			Template: cp.PodTemplateSpec,
		},
	}
	statefulSet.Labels[useTemplatesLabel] = "true"
	statefulSet.Spec.Selector.MatchLabels[useTemplatesLabel] = "true"
	statefulSet.Spec.Template.Labels[useTemplatesLabel] = "true"

	if service.Spec.UpdateStrategy == "on-delete" {
		statefulSet.Spec.UpdateStrategy.Type = appsv1.OnDeleteStatefulSetStrategyType
	} else {
		zero := int32(0)
		statefulSet.Spec.UpdateStrategy.Type = appsv1.RollingUpdateStatefulSetStrategyType
		statefulSet.Spec.UpdateStrategy.RollingUpdate = &appsv1.RollingUpdateStatefulSetStrategy{
			Partition: &zero,
		}
	}

	if service.Spec.DeploymentStrategy == "ordered" {
		statefulSet.Spec.PodManagementPolicy = appsv1.OrderedReadyPodManagement
	} else {
		statefulSet.Spec.PodManagementPolicy = appsv1.RollingUpdateStatefulSetStrategyType
	}

	for _, volumeTemplate := range usedTemplates {
		labels := map[string]string{
			"rio.cattle.io/stack":          stack.Stack.Name,
			"rio.cattle.io/workspace":      stack.Stack.Namespace,
			"rio.cattle.io/volumetemplate": volumeTemplate.Name,
		}

		pvc, err := volume.ToPVC(stack.Namespace, labels, *volumeTemplate)
		if err != nil {
			return err
		}

		statefulSet.Spec.VolumeClaimTemplates = append(statefulSet.Spec.VolumeClaimTemplates, *pvc)
	}

	output.StatefulSets[statefulSet.Name] = statefulSet
	return nil
}
