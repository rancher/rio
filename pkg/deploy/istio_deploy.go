package deploy

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/rancher/norman/types"
	"github.com/rancher/rio/pkg/settings"
	"istio.io/istio/pilot/pkg/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type IstioObject struct {
	types.Namespaced

	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec interface{} `json:"spec,omitempty"`
}

func (i *IstioObject) DeepCopyObject() runtime.Object {
	panic("not implemented")
}

func IstioObjects(namespace string, stack *StackResources) ([]runtime.Object, error) {
	if settings.ClusterDomain.Get() == "" {
		return nil, fmt.Errorf("waiting for cluster domain")
	}

	ds := destinations(nil, stack)

	vs, err := virtualservices(nil, stack)
	if err != nil {
		return nil, err
	}

	if len(vs) == 0 {
		return nil, nil
	}

	return convertIstioObjects(append(ds, vs...))
}

func convertIstioObjects(objs []runtime.Object) ([]runtime.Object, error) {
	for _, obj := range objs {
		if istioObject, ok := obj.(*IstioObject); ok {
			if pb, ok := istioObject.Spec.(proto.Message); ok {
				m, err := model.ToJSONMap(pb)
				if err != nil {
					return nil, err
				}
				istioObject.Spec = m
			}
		}
	}

	return objs, nil
}
