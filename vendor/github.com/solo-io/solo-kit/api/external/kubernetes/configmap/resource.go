package configmap

import (
	"reflect"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	kubev1 "k8s.io/api/core/v1"
)

type ConfigMap kubev1.ConfigMap

var _ resources.Resource = new(ConfigMap)

func (p *ConfigMap) Clone() *ConfigMap {
	vp := kubev1.ConfigMap(*p)
	copy := vp.DeepCopy()
	newP := ConfigMap(*copy)
	return &newP
}

func (p *ConfigMap) GetMetadata() core.Metadata {
	return kubeutils.FromKubeMeta(p.ObjectMeta)
}

func (p *ConfigMap) SetMetadata(meta core.Metadata) {
	p.ObjectMeta = kubeutils.ToKubeMeta(meta)
}

func (p *ConfigMap) Equal(that interface{}) bool {
	return reflect.DeepEqual(p, that)
}
