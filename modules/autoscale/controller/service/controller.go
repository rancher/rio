package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/knative/serving/pkg/apis/autoscaling/v1alpha1"
	"github.com/knative/serving/pkg/apis/networking"
	servingv1beta1 "github.com/knative/serving/pkg/apis/serving/v1beta1"
	autoscalev1 "github.com/rancher/rio/pkg/apis/autoscale.rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
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

const (
	//TODO Keep in sync with vendor/github.com/knative/serving/pkg/apis/autoscaling?
	GroupName = "autoscaling.knative.dev"

	// MinScaleAnnotationKey is the annotation to specify the minimum number of Pods
	// the PodAutoscaler should provision. For example,
	//   autoscaling.knative.dev/minScale: "1"
	MinScaleAnnotationKey = GroupName + "/minScale"
	// MaxScaleAnnotationKey is the annotation to specify the maximum number of Pods
	// the PodAutoscaler should provision. For example,
	//   autoscaling.knative.dev/maxScale: "10"
	MaxScaleAnnotationKey = GroupName + "/maxScale"

	ConfigurationKey = "serving.knative.dev/configuration"

	ServiceKey = "serving.knative.dev/service"

	RevisionKey = "serving.knative.dev/revision"

	ScrapeKey = "metric-scrape"

	ReferingLabel = "autoscaling.knative.dev/class"

	ContainerPortKey = "container-port"

	AppKey = "app"

	VersionKey = "version"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "autoscale-service", rContext.Rio.Rio().V1().Service())
	c.Apply = c.Apply.WithCacheTypes(rContext.AutoScale.Autoscale().V1().ServiceScaleRecommendation(),
		rContext.Serving.Autoscaling().V1alpha1().PodAutoscaler())

	p := populator{
		systemNamespace: rContext.Namespace,
	}
	c.Populator = p.populate

	h := handler{
		services: rContext.Rio.Rio().V1().Service(),
	}
	rContext.AutoScale.Autoscale().V1().ServiceScaleRecommendation().OnChange(ctx, "ssr-service-update", h.sync)

	return nil
}

type populator struct {
	systemNamespace string
}

func (p populator) populate(object runtime.Object, ns *corev1.Namespace, os *objectset.ObjectSet) error {
	populatePodAutoscaler(object, ns, os)
	populateServiceRecommendation(object, ns, os)
	return nil
}

func populateServiceRecommendation(object runtime.Object, ns *corev1.Namespace, os *objectset.ObjectSet) {
	service := object.(*v1.Service)
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
	if svc.Status.ObservedScale != nil && *svc.Status.ObservedScale == observedScale {
		return nil
	}
	logrus.Infof("Setting desired scale %v for %v/%v", *ssr.Spec.Replicas, svc.Namespace, svc.Name)

	svc.Status.ObservedScale = &observedScale
	if _, err := h.services.Update(svc); err != nil {
		return err
	}
	return nil
}

func populatePodAutoscaler(object runtime.Object, ns *corev1.Namespace, os *objectset.ObjectSet) {
	service := object.(*v1.Service)
	autoscale := AutoscaleEnabled(service)
	if !autoscale {
		return
	}
	app, version := services.AppAndVersion(service)
	annotation := map[string]string{
		ReferingLabel:         "kpa.autoscaling.knative.dev",
		MinScaleAnnotationKey: strconv.Itoa(*service.Spec.Autoscale.MinReplicas),
		MaxScaleAnnotationKey: strconv.Itoa(*service.Spec.Autoscale.MaxReplicas),
	}
	if constants.ServiceMeshMode == constants.ServiceMeshModeLinkerd {
		annotation[ScrapeKey] = "linkerd"
	} else if constants.ServiceMeshMode == constants.ServiceMeshModeIstio {
		annotation[ScrapeKey] = "envoy"
	}
	var portValue string
	for _, port := range service.Spec.Ports {
		if port.Expose != nil && *port.Expose {
			portValue = strconv.Itoa(int(port.TargetPort))
			break
		}
	}
	podAutoscaler := constructors.NewPodAutoscaler(service.Namespace, service.Name, v1alpha1.PodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: annotation,
			Labels: map[string]string{
				ConfigurationKey: service.Name,
				ServiceKey:       service.Name,
				RevisionKey:      fmt.Sprintf("%s-%s", app, version),
				ContainerPortKey: portValue,
				AppKey:           app,
				VersionKey:       version,
			},
		},
		Spec: v1alpha1.PodAutoscalerSpec{
			ContainerConcurrency: servingv1beta1.RevisionContainerConcurrencyType(*service.Spec.Autoscale.Concurrency),
			ScaleTargetRef: corev1.ObjectReference{
				Kind:       "ServiceScaleRecommendation",
				APIVersion: autoscalev1.SchemeGroupVersion.String(),
				Name:       service.Name,
			},
			ProtocolType: networking.ProtocolHTTP1,
		},
	})

	os.Add(podAutoscaler)
	return
}

func AutoscaleEnabled(service *v1.Service) bool {
	return service.Spec.Autoscale.MinReplicas != nil && service.Spec.Autoscale.MaxReplicas != nil &&
		service.Spec.Autoscale.Concurrency != nil && *service.Spec.Autoscale.MinReplicas != *service.Spec.Autoscale.MaxReplicas
}
