package monitoring

import (
	"github.com/rancher/rio/pkg/apply"
	"github.com/rancher/rio/pkg/data"
	"github.com/rancher/rio/pkg/settings"
	projectv1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Reconcile(feature *projectv1.Feature) error {
	monitoringStack := make([]runtime.Object, 0)
	if feature.Spec.Enable {
		monitoringStack = append(monitoringStack, data.Stack("istio-telemetry", v1.StackSpec{
			DisableMesh: true,
			Answers: map[string]string{
				"LB_NAMESPACE": settings.IstioExternalLBNamespace,
			},
			EnableKubernetesResources: true,
		}))
	}
	empty := []string{}
	if len(monitoringStack) == 0 {
		empty = []string{"stacks.rio.cattle.io"}
	}
	return apply.Apply(monitoringStack, empty, settings.RioSystemNamespace, "rio-monitoring-stacks")
}
