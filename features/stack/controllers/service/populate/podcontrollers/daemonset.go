package podcontrollers

import (
	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/pkg/namespace"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	appsv1 "k8s.io/api/apps/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func daemonSet(stack *riov1.Stack, service *riov1.Service, cp *controllerParams, os *objectset.ObjectSet) {
	ns, name := namespace.NameRefWithNamespace(service.Name, stack)
	daemonSet := &appsv1.DaemonSet{
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
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: cp.SelectorLabels,
			},
			Template: cp.PodTemplateSpec,
		},
	}

	if service.Spec.UpdateStrategy == "on-delete" {
		daemonSet.Spec.UpdateStrategy.Type = appsv1.OnDeleteStatefulSetStrategyType
	} else {
		daemonSet.Spec.UpdateStrategy.Type = appsv1.RollingUpdateStatefulSetStrategyType
		if cp.Scale.Scale > 0 && cp.Scale.BatchSize > 0 {
			daemonSet.Spec.UpdateStrategy.RollingUpdate = &appsv1.RollingUpdateDaemonSet{
				MaxUnavailable: cp.Scale.MaxUnavailable,
			}
		}
	}

	os.Add(daemonSet)
}
