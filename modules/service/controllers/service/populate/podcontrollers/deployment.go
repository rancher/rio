package podcontrollers

import (
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deployment(service *riov1.Service, cp *controllerParams, os *objectset.ObjectSet) {
	dep := constructors.NewDeployment(service.Namespace, service.Name, appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      cp.Labels,
			Annotations: map[string]string{},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &cp.Scale.Scale,
			Selector: &metav1.LabelSelector{
				MatchLabels: cp.SelectorLabels,
			},
			Template: cp.PodTemplateSpec,
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
		},
	})

	if cp.Scale.Scale > 0 && cp.Scale.BatchSize > 0 {
		dep.Spec.Strategy.RollingUpdate = &appsv1.RollingUpdateDeployment{
			MaxSurge:       cp.Scale.MaxSurge,
			MaxUnavailable: cp.Scale.MaxUnavailable,
		}
	}

	os.Add(dep)
}

func deploymentForSystem(service *riov1.Service, cp *controllerParams, os *objectset.ObjectSet) {

}
