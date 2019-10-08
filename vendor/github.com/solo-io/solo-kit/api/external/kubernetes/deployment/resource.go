package deployment

import (
	"reflect"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	v1 "k8s.io/api/apps/v1"
)

type Deployment v1.Deployment

var _ resources.Resource = new(Deployment)

func (p *Deployment) Clone() *Deployment {
	vp := v1.Deployment(*p)
	copy := vp.DeepCopy()
	newP := Deployment(*copy)
	return &newP
}

func (p *Deployment) GetMetadata() core.Metadata {
	return kubeutils.FromKubeMeta(p.ObjectMeta)
}

func (p *Deployment) SetMetadata(meta core.Metadata) {
	p.ObjectMeta = kubeutils.ToKubeMeta(meta)
}

func (p *Deployment) Equal(that interface{}) bool {
	return reflect.DeepEqual(p, that)
}
