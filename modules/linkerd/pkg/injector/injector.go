package injector

import (
	"github.com/rancher/wrangler/pkg/apply/injectors"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func RegisterInjector() {
	injectors.Register("mesh", addLinkerdLabel)
}

func addLinkerdLabel(objs []runtime.Object) ([]runtime.Object, error) {
	for _, obj := range objs {
		switch o := obj.(type) {
		case *appsv1.Deployment:
			setAnnotations(&o.ObjectMeta)
		case *appsv1.StatefulSet:
			setAnnotations(&o.ObjectMeta)
		case *appsv1.DaemonSet:
			setAnnotations(&o.ObjectMeta)
		}
	}
	return objs, nil
}

func setAnnotations(meta *v1.ObjectMeta) {
	if meta.Annotations["rio.cattle.io/mesh"] != "true" {
		return
	}
	if meta.Annotations == nil {
		meta.Annotations = map[string]string{}
	}
	meta.Annotations["linkerd.io/inject"] = "enabled"
}
