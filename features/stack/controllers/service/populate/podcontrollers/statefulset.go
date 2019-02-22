package podcontrollers

import (
	"github.com/rancher/norman/pkg/objectset"
	populate2 "github.com/rancher/rio/features/stack/controllers/volume/populate"
	"github.com/rancher/rio/pkg/namespace"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	appsv1 "k8s.io/api/apps/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const useTemplatesLabel = "rio.cattle.io/use-templates"

func statefulSet(stack *riov1.Stack, service *riov1.Service, cp *controllerParams, usedTemplates map[string]*riov1.Volume, os *objectset.ObjectSet) error {
	ns, name := namespace.NameRefWithNamespace(service.Name, stack)
	statefulSet := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1beta2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   ns,
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
			"rio.cattle.io/stack":           stack.Name,
			"rio.cattle.io/project":         stack.Namespace,
			"rio.cattle.io/volume-template": volumeTemplate.Name,
		}

		pvc, err := populate2.ToPVC(labels, *volumeTemplate, stack)
		if err != nil {
			return err
		}

		statefulSet.Spec.VolumeClaimTemplates = append(statefulSet.Spec.VolumeClaimTemplates, *pvc)
	}

	os.Add(statefulSet)
	return nil
}
