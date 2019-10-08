package service

import (
	"reflect"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	kubev1 "k8s.io/api/core/v1"
)

type Service kubev1.Service

var _ resources.Resource = new(Service)

func (p *Service) Clone() *Service {
	vp := kubev1.Service(*p)
	copy := vp.DeepCopy()
	newP := Service(*copy)
	return &newP
}

func (p *Service) GetMetadata() core.Metadata {
	return kubeutils.FromKubeMeta(p.ObjectMeta)
}

func (p *Service) SetMetadata(meta core.Metadata) {
	p.ObjectMeta = kubeutils.ToKubeMeta(meta)
}

func (p *Service) Equal(that interface{}) bool {
	return reflect.DeepEqual(p, that)
}
