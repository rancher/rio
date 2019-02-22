package autoscale

import (
	"github.com/rancher/norman/pkg/objectset"
	v12 "github.com/rancher/rio/types/apis/rio-autoscale.cattle.io/v1"
	v1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Populate(stack *v1.Stack, services []*v1.Service, os *objectset.ObjectSet) {
	labels := map[string]string{}
	labels["rio.cattle.io/stack"] = stack.Name
	labels["rio.cattle.io/project"] = stack.Namespace
	for _, s := range services {
		if s.Spec.AutoScale != nil {
			spec := v12.ServiceScaleRecommendation{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v12.ServiceScaleRecommendationSpec{
					MinScale:          int32(s.Spec.AutoScale.MinScale),
					MaxScale:          int32(s.Spec.AutoScale.MaxScale),
					Concurrency:       s.Spec.AutoScale.Concurrency,
					PrometheusURL:     "http://prometheus.prometheus:9090",
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
