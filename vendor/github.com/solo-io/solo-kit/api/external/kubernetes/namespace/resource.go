package namespace

import (
	"reflect"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	kubev1 "k8s.io/api/core/v1"
)

type KubeNamespace kubev1.Namespace

func (p *KubeNamespace) Clone() *KubeNamespace {
	vp := kubev1.Namespace(*p)
	copy := vp.DeepCopy()
	newP := KubeNamespace(*copy)
	return &newP
}

func (p *KubeNamespace) GetMetadata() core.Metadata {
	return kubeutils.FromKubeMeta(p.ObjectMeta)
}

func (p *KubeNamespace) SetMetadata(meta core.Metadata) {
	p.ObjectMeta = kubeutils.ToKubeMeta(meta)
}

func (p *KubeNamespace) Equal(that interface{}) bool {
	return reflect.DeepEqual(p, that)
}
