package service

import (
	"context"
	"fmt"

	autoscalev1 "github.com/rancher/rio/pkg/apis/autoscale.rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "autoscale-service", rContext.Rio.Rio().V1().Service())
	c.Apply = c.Apply.WithCacheTypes(rContext.AutoScale.Autoscale().V1().ServiceScaleRecommendation())

	p := populator{
		systemNamespace: rContext.Namespace,
	}
	c.Populator = p.populateServiceRecommendation

	return nil
}

type populator struct {
	systemNamespace string
}

func (p populator) populateServiceRecommendation(object runtime.Object, ns *corev1.Namespace, os *objectset.ObjectSet) error {
	service := object.(*v1.Service)
	labels := map[string]string{}
	autoscale := false
	if service.Spec.MinScale != nil && service.Spec.MaxScale != nil && service.Spec.Concurrency != nil {
		autoscale = true
	}
	if autoscale {
		spec := autoscalev1.ServiceScaleRecommendation{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: autoscalev1.ServiceScaleRecommendationSpec{
				MinScale:          int32(*service.Spec.MinScale),
				MaxScale:          int32(*service.Spec.MaxScale),
				Concurrency:       *service.Spec.Concurrency,
				PrometheusURL:     fmt.Sprintf("http://prometheus.%s:9090", p.systemNamespace),
				ServiceNameToRead: service.Name,
			},
			Status: autoscalev1.ServiceScaleRecommendationStatus{
				DesiredScale: &[]int32{int32(service.Spec.Scale)}[0],
			},
		}
		ssr := autoscalev1.NewServiceScaleRecommendation(service.Namespace, service.Name, spec)
		autoscalev1.ServiceScaleRecommendationSynced.True(ssr)
		os.Add(ssr)
	}
	return nil
}
