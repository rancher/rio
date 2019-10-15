package service

import (
	"context"

	autoscalev1 "github.com/rancher/rio/pkg/apis/autoscale.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	autoscalev1controller "github.com/rancher/rio/pkg/generated/controllers/autoscale.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		services: rContext.Rio.Rio().V1().Service(),
	}
	riov1controller.RegisterServiceGeneratingHandler(
		ctx,
		rContext.Rio.Rio().V1().Service(),
		rContext.Apply.WithCacheTypes(rContext.AutoScale.Autoscale().V1().ServiceScaleRecommendation()),
		"ServiceRecommendationGenerated",
		"serviceRecommendation",
		h.populate,
		nil,
	)

	rContext.AutoScale.Autoscale().V1().ServiceScaleRecommendation().OnChange(ctx, "ssr-service-update", h.sync)

	return nil
}

func (h handler) populate(service *riov1.Service, status riov1.ServiceStatus) ([]runtime.Object, riov1.ServiceStatus, error) {
	os := objectset.NewObjectSet()
	populateServiceRecommendation(service, os)
	return os.All(), status, nil
}

func populateServiceRecommendation(service *riov1.Service, os *objectset.ObjectSet) {
	autoscale := AutoscaleEnabled(service)
	app, version := services.AppAndVersion(service)
	if autoscale {
		spec := autoscalev1.ServiceScaleRecommendation{
			Spec: autoscalev1.ServiceScaleRecommendationSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app":     app,
						"version": version,
					},
				},
				MinScale:    *service.Spec.Autoscale.MinReplicas,
				MaxScale:    *service.Spec.Autoscale.MaxReplicas,
				Concurrency: service.Spec.Autoscale.Concurrency,
			},
			Status: autoscalev1.ServiceScaleRecommendationStatus{},
		}
		ssr := autoscalev1.NewServiceScaleRecommendation(service.Namespace, service.Name, spec)
		autoscalev1.ServiceScaleRecommendationSynced.True(ssr)
		os.Add(ssr)
	}
	return
}

type handler struct {
	services riov1controller.ServiceController
	ssrs     autoscalev1controller.ServiceScaleRecommendationController
}

func (h handler) sync(key string, obj *autoscalev1.ServiceScaleRecommendation) (*autoscalev1.ServiceScaleRecommendation, error) {
	if obj == nil || obj.DeletionTimestamp != nil {
		return obj, nil
	}
	return obj, h.setServiceScale(obj)
}

func (h handler) setServiceScale(ssr *autoscalev1.ServiceScaleRecommendation) error {
	svc, err := h.services.Cache().Get(ssr.Namespace, ssr.Name)
	if err != nil {
		return err
	}
	if ssr.Spec.Replicas == nil {
		return nil
	}

	observedScale := int(*ssr.Spec.Replicas)
	if svc.Status.ComputedReplicas != nil && *svc.Status.ComputedReplicas == observedScale {
		return nil
	}
	logrus.Infof("Setting desired scale %v for %v/%v", *ssr.Spec.Replicas, svc.Namespace, svc.Name)

	svc.Status.ComputedReplicas = &observedScale
	if _, err := h.services.Update(svc); err != nil {
		return err
	}
	return nil
}

func AutoscaleEnabled(service *v1.Service) bool {
	return service.Spec.Autoscale != nil && service.Spec.Autoscale.MinReplicas != nil && service.Spec.Autoscale.MaxReplicas != nil && *service.Spec.Autoscale.MinReplicas != *service.Spec.Autoscale.MaxReplicas
}
