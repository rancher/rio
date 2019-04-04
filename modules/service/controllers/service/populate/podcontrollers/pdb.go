package podcontrollers

import (
	"fmt"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/objectset"
	v1beta12 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func pdb(service *riov1.Service, cp *controllerParams, os *objectset.ObjectSet) {
	if !(cp.Scale.Scale > 0 && cp.Scale.BatchSize > 0 && cp.Scale.BatchSize < int(cp.Scale.Scale)) {
		return
	}

	if service.Spec.AutoScale != nil {
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
			Namespace: service.Namespace,
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

	os.Add(pdb)
}
