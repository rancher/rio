package injector

import (
	"github.com/rancher/wrangler/pkg/apply/injectors"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1/kube/apis/gloo.solo.io/v1"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
)

func RegisterInjector() {
	injectors.Register("gloo-mesh", injectLinkerd)
}

func injectLinkerd(objs []runtime.Object) ([]runtime.Object, error) {
	for i, obj := range objs {
		switch o := obj.(type) {
		case *unstructured.Unstructured:
			if o.GetKind() == "Deployment" {
				data, err := o.MarshalJSON()
				if err != nil {
					return nil, err
				}
				deploy := &v1.Deployment{}
				if err := json.Unmarshal(data, deploy); err != nil {
					return nil, err
				}
				if deploy.Spec.Template.Annotations == nil {
					deploy.Spec.Template.Annotations = map[string]string{}
				}
				deploy.Spec.Template.Annotations["linkerd.io/inject"] = "enabled"
				objs[i] = deploy
			} else if o.GetKind() == "Settings" {
				data, err := o.MarshalJSON()
				if err != nil {
					return nil, err
				}
				setting := &gloov1.Settings{}
				if err := json.Unmarshal(data, setting); err != nil {
					return nil, err
				}
				setting.Spec.Linkerd = true
				objs[i] = setting
			}
		}
	}
	return objs, nil
}
