package autoscale

import (
	"github.com/rancher/norman/pkg/objectset"
	v12 "github.com/rancher/rio/types/apis/rio-autoscale.cattle.io/v1"
	v1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

func Populate(services []*v1.Service, os *objectset.ObjectSet) {
	for _, s := range services {
		if s.Spec.EnableAutoScale {
			spec := v12.ServiceScaleRecommendation{
				Spec: v12.ServiceScaleRecommendationSpec{
					MinScale:          int32(s.Spec.AutoscaleConfig.MinScale),
					MaxScale:          int32(s.Spec.AutoscaleConfig.MaxScale),
					Concurrency:       s.Spec.AutoscaleConfig.ContainerConcurrency,
					PrometheusURL:     "http://prometheus",
					ServiceNameToRead: s.Name,
				},
				Status: v12.ServiceScaleRecommendationStatus{
					DesiredScale: &[]int32{int32(s.Spec.Scale)}[0],
				},
			}
			ssr := v12.NewServiceScaleRecommendation(s.Namespace, s.Name, spec)
			v12.ServiceScaleRecommendationSynced.True(ssr)
			os.Add(ssr)
		}
	}
}
