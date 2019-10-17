package injector

import (
	"github.com/rancher/wrangler/pkg/apply/injectors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func RegisterInjector() {
	injectors.Register("mesh", addLinkerdLabel)
}

func addLinkerdLabel(objs []runtime.Object) ([]runtime.Object, error) {
	for _, obj := range objs {
		switch o := obj.(type) {
		case *appsv1.Deployment:
			setAnnotations(o.Spec.Template.Annotations)
		case *appsv1.StatefulSet:
			setAnnotations(o.Spec.Template.Annotations)
		case *appsv1.DaemonSet:
			setAnnotations(o.Spec.Template.Annotations)
		}
	}
	return objs, nil
}

func setAnnotations(annotation map[string]string) {
	if annotation["rio.cattle.io/mesh"] != "true" {
		return
	}
	if annotation == nil {
		annotation = map[string]string{}
	}
	annotation["linkerd.io/inject"] = "enabled"
}
