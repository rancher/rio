package populate

import (
	autoscalev1 "github.com/rancher/rio/pkg/apis/autoscale.rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/objectset"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func ServiceRecommendationForService(object runtime.Object, ns *corev1.Namespace, os *objectset.ObjectSet) error {
	service := object.(*v1.Service)
	labels := map[string]string{}
	autoscale := service.Spec.MinScale != service.Spec.MaxScale
	if autoscale {
		spec := autoscalev1.ServiceScaleRecommendation{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: autoscalev1.ServiceScaleRecommendationSpec{
				MinScale:          int32(service.Spec.MinScale),
				MaxScale:          int32(service.Spec.MaxScale),
				Concurrency:       service.Spec.Concurrency,
				PrometheusURL:     "http://prometheus.prometheus:9090",
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
