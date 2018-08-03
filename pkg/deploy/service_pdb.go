package deploy

import (
	"fmt"

	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	v1beta12 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func pdbs(objects []runtime.Object, name, namespace string, matchLabels map[string]string, service *v1beta1.ServiceUnversionedSpec) []runtime.Object {
	scaleParams := parseScaleParams(service)
	if !(scaleParams.Scale > 0 && scaleParams.BatchSize > 0 && scaleParams.BatchSize < int(scaleParams.Scale)) {
		return objects
	}

	pdbSize := service.BatchSize
	if service.BatchSize > service.Scale {
		pdbSize = 1
	}
	pdbQuantity := intstr.FromInt(pdbSize)

	pdb := &v1beta12.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%d", name, pdbQuantity.IntVal),
			Namespace: namespace,
		},
		Spec: v1beta12.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: matchLabels,
			},
			MaxUnavailable: &pdbQuantity,
		},
		Status: v1beta12.PodDisruptionBudgetStatus{
			DisruptedPods: map[string]metav1.Time{},
		},
	}

	objects = append(objects, pdb)
	return objects
}
