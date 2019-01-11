package autoscale

import (
	"fmt"

	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/pkg/settings"
	v12 "github.com/rancher/rio/types/apis/rio-autoscale.cattle.io/v1"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

func Populate(services []*v1.Service, os *objectset.ObjectSet) {
	for _, s := range services {
		if s.Spec.EnableAutoScale {
			spec := v12.ServiceScaleRecommendation{
				Spec: v12.ServiceScaleRecommendationSpec{
					MinScale:          int32(s.Spec.AutoscaleConfig.MinScale),
					MaxScale:          int32(s.Spec.AutoscaleConfig.MaxScale),
					Concurrency:       s.Spec.AutoscaleConfig.ContainerConcurrency,
					ZeroScaleService:  fmt.Sprintf("gateway.%s.svc.cluster.local", namespace.StackNamespace(settings.RioSystemNamespace, settings.AutoScaleStack)),
					PrometheusURL:     fmt.Sprintf("http://prometheus.%s.svc.cluster.local:9090", namespace.StackNamespace(settings.RioSystemNamespace, settings.Prometheus)),
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
