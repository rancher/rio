package service

import (
	"reflect"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	kubev1 "k8s.io/api/core/v1"
)

type Service struct {
	kubev1.Service
	cachedMeta *core.Metadata
}

var _ resources.Resource = new(Service)

func NewService(kubeSvc kubev1.Service) *Service {
	meta := kubeutils.FromKubeMeta(kubeSvc.ObjectMeta)
	var s Service
	s.cachedMeta = &meta
	s.Service = kubeSvc
	return &s
}
func (p *Service) GetKubeService() kubev1.Service {
	return p.Service
}

func (p *Service) Clone() *Service {
	vp := kubev1.Service(p.Service)
	copy := vp.DeepCopy()
	newP := Service{Service: *copy}
	return &newP
}

func (p *Service) GetMetadata() core.Metadata {
	if p.cachedMeta == nil {
		meta := kubeutils.FromKubeMeta(p.ObjectMeta)
		p.cachedMeta = &meta
	}
	return *p.cachedMeta
}

func (p *Service) SetMetadata(meta core.Metadata) {
	p.ObjectMeta = kubeutils.ToKubeMeta(meta)
	// copy so we own everything
	meta = kubeutils.FromKubeMeta(p.ObjectMeta)
	p.cachedMeta = &meta
}

func (p *Service) Equal(that interface{}) bool {
	p2, ok := that.(*Service)
	if !ok {
		return false
	}
	return reflect.DeepEqual(p.Service, p2.Service)
}
