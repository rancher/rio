package service

import (
	"context"
	"fmt"
	"time"

	autoscalev1 "github.com/rancher/rio/pkg/apis/autoscale.rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	autoscalev1controller "github.com/rancher/rio/pkg/generated/controllers/autoscale.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	"github.com/sirupsen/logrus"
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

	h := handler{
		services: rContext.Rio.Rio().V1().Service(),
	}
	rContext.AutoScale.Autoscale().V1().ServiceScaleRecommendation().OnChange(ctx, "ssr-service-update", h.sync)

	return nil
}

type populator struct {
	systemNamespace string
}

func (p populator) populateServiceRecommendation(object runtime.Object, ns *corev1.Namespace, os *objectset.ObjectSet) error {
	service := object.(*v1.Service)
	labels := map[string]string{}
	autoscale := false
	if service.Spec.MinScale != nil && service.Spec.MaxScale != nil && service.Spec.Concurrency != nil && *service.Spec.MinScale != *service.Spec.MaxScale {
		autoscale = true
	}
	app, version := services.AppAndVersion(service)
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
				Selector: map[string]string{
					"app":     app,
					"version": version,
				},
			},
			Status: autoscalev1.ServiceScaleRecommendationStatus{},
		}
		ssr := autoscalev1.NewServiceScaleRecommendation(service.Namespace, service.Name, spec)
		autoscalev1.ServiceScaleRecommendationSynced.True(ssr)
		os.Add(ssr)
	}
	return nil
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
	// wait for a minute after scale from zero
	if svc.Status.ScaleFromZeroTimestamp != nil && svc.Status.ScaleFromZeroTimestamp.Add(time.Minute).After(time.Now()) {
		logrus.Infof("skipping setting scale because service  %s/%s is scaled from zero within a minute", svc.Namespace, svc.Name)
		go func() {
			time.Sleep(time.Second * 60)
			h.ssrs.Enqueue(ssr.Namespace, ssr.Name)
		}()
		return nil
	}

	if ssr.Status.DesiredScale == nil {
		return nil
	}
	observedScale := int(*ssr.Status.DesiredScale)
	if svc.Status.ObservedScale != nil && *svc.Status.ObservedScale == observedScale {
		return nil
	}
	logrus.Infof("Setting desired scale %v for %v/%v", *ssr.Status.DesiredScale, svc.Namespace, svc.Name)

	svc.Status.ObservedScale = &observedScale
	if _, err := h.services.Update(svc); err != nil {
		return err
	}
	return nil
}
