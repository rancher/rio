package customresourcedefinition

import (
	"reflect"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

type CustomResourceDefinition v1beta1.CustomResourceDefinition

var _ resources.Resource = new(CustomResourceDefinition)

func (p *CustomResourceDefinition) Clone() *CustomResourceDefinition {
	vp := v1beta1.CustomResourceDefinition(*p)
	copy := vp.DeepCopy()
	newP := CustomResourceDefinition(*copy)
	return &newP
}

func (p *CustomResourceDefinition) GetMetadata() core.Metadata {
	return kubeutils.FromKubeMeta(p.ObjectMeta)
}

func (p *CustomResourceDefinition) SetMetadata(meta core.Metadata) {
	p.ObjectMeta = kubeutils.ToKubeMeta(meta)
}

func (p *CustomResourceDefinition) Equal(that interface{}) bool {
	return reflect.DeepEqual(p, that)
}
