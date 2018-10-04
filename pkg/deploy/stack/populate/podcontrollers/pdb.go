package podcontrollers

import (
	"fmt"

	"github.com/rancher/rio/pkg/deploy/stack/input"
	"github.com/rancher/rio/pkg/deploy/stack/output"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	v1beta12 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func pdb(stack *input.Stack, service *v1beta1.Service, cp *controllerParams, output *output.Deployment) {
	if !(cp.Scale.Scale > 0 && cp.Scale.BatchSize > 0 && cp.Scale.BatchSize < int(cp.Scale.Scale)) {
		return
	}

	pdbSize := service.Spec.BatchSize
	if service.Spec.BatchSize > service.Spec.Scale {
		pdbSize = 1
	}
	pdbQuantity := intstr.FromInt(pdbSize)

	pdb := &v1beta12.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%d", service.Name, pdbQuantity.IntVal),
			Namespace: stack.Namespace,
		},
		Spec: v1beta12.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: cp.SelectorLabels,
			},
			MaxUnavailable: &pdbQuantity,
		},
		Status: v1beta12.PodDisruptionBudgetStatus{
			DisruptedPods: map[string]metav1.Time{},
		},
	}

	output.PodDisruptionBudgets[pdb.Name] = pdb
}
