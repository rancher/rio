package job

import (
	"reflect"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	batchv1 "k8s.io/api/batch/v1"
)

type Job batchv1.Job

var _ resources.Resource = new(Job)

func (p *Job) Clone() *Job {
	vp := batchv1.Job(*p)
	copy := vp.DeepCopy()
	newP := Job(*copy)
	return &newP
}

func (p *Job) GetMetadata() core.Metadata {
	return kubeutils.FromKubeMeta(p.ObjectMeta)
}

func (p *Job) SetMetadata(meta core.Metadata) {
	p.ObjectMeta = kubeutils.ToKubeMeta(meta)
}

func (p *Job) Equal(that interface{}) bool {
	return reflect.DeepEqual(p, that)
}
