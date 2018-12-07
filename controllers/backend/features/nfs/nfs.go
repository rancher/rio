package nfs

import (
	"github.com/rancher/rio/pkg/apply"
	"github.com/rancher/rio/pkg/data"
	"github.com/rancher/rio/pkg/settings"
	projectv1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Reconcile(feature *projectv1.Feature) error {
	var result []runtime.Object
	if feature.Spec.Enable {
		result = append(result, data.Stack("nfs", v1.StackSpec{
			DisableMesh: true,
			Answers:     feature.Spec.Answers,
		}))
	}
	empty := []string{}
	if len(result) == 0 {
		empty = []string{"stacks.rio.cattle.io"}
	}

	return apply.Apply(result, empty, settings.RioSystemNamespace, "rio-nfs-stacks")
}
