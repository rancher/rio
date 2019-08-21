package service

import (
	"strconv"
	"testing"

	"github.com/knative/serving/pkg/apis/autoscaling/v1alpha1"
	"github.com/knative/serving/pkg/apis/networking"
	servingv1beta1 "github.com/knative/serving/pkg/apis/serving/v1beta1"
	"github.com/rancher/rio/modules/test"
	autoscalev1 "github.com/rancher/rio/pkg/apis/autoscale.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPopulateServiceRecommendationMinMaxNotEqual(t *testing.T) {
	os := objectset.NewObjectSet()
	input := riov1.NewService("default", "test", riov1.Service{
		Spec: riov1.ServiceSpec{
			AutoscaleConfig: riov1.AutoscaleConfig{
				MinScale:    &[]int{0}[0],
				MaxScale:    &[]int{10}[0],
				Concurrency: &[]int{10}[0],
			},
			ServiceRevision: riov1.ServiceRevision{
				App:     "foo",
				Version: "v0",
			},
		},
	})

	expected := autoscalev1.NewServiceScaleRecommendation(input.Namespace, input.Name, autoscalev1.ServiceScaleRecommendation{
		Spec: autoscalev1.ServiceScaleRecommendationSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":     "foo",
					"version": "v0",
				},
			},
		},
	})
	autoscalev1.ServiceScaleRecommendationSynced.True(expected)

	populateServiceRecommendation(input, nil, os)

	test.AssertObjects(t, expected, os)
}

func TestPopulateServiceRecommendationMinMaxEqual(t *testing.T) {
	os := objectset.NewObjectSet()
	service := riov1.NewService("default", "test", riov1.Service{
		Spec: riov1.ServiceSpec{
			AutoscaleConfig: riov1.AutoscaleConfig{
				MinScale:    &[]int{10}[0],
				MaxScale:    &[]int{10}[0],
				Concurrency: &[]int{10}[0],
			},
			ServiceRevision: riov1.ServiceRevision{
				App:     "foo",
				Version: "v0",
			},
		},
	})

	populateServiceRecommendation(service, nil, os)

	assert.Assert(t, os.Len() == 0, "Should not populate one serviceScaleRecommendation object")
}

func TestPopulateServiceRecommendationMinNil(t *testing.T) {
	os := objectset.NewObjectSet()
	service := riov1.NewService("default", "test", riov1.Service{
		Spec: riov1.ServiceSpec{
			AutoscaleConfig: riov1.AutoscaleConfig{
				MaxScale:    &[]int{10}[0],
				Concurrency: &[]int{10}[0],
			},
			ServiceRevision: riov1.ServiceRevision{
				App:     "foo",
				Version: "v0",
			},
		},
	})

	populateServiceRecommendation(service, nil, os)

	assert.Assert(t, os.Len() == 0, "Should not populate one serviceScaleRecommendation object")
}

func TestPopulateServiceRecommendationMaxNil(t *testing.T) {
	os := objectset.NewObjectSet()
	service := riov1.NewService("default", "test", riov1.Service{
		Spec: riov1.ServiceSpec{
			AutoscaleConfig: riov1.AutoscaleConfig{
				MinScale:    &[]int{10}[0],
				Concurrency: &[]int{10}[0],
			},
			ServiceRevision: riov1.ServiceRevision{
				App:     "foo",
				Version: "v0",
			},
		},
	})

	populateServiceRecommendation(service, nil, os)

	assert.Assert(t, os.Len() == 0, "Should not populate one serviceScaleRecommendation object")
}

func TestPopulateServiceRecommendationMinMaxNil(t *testing.T) {
	os := objectset.NewObjectSet()
	service := riov1.NewService("default", "test", riov1.Service{
		Spec: riov1.ServiceSpec{
			AutoscaleConfig: riov1.AutoscaleConfig{
				Concurrency: &[]int{10}[0],
			},
			ServiceRevision: riov1.ServiceRevision{
				App:     "foo",
				Version: "v0",
			},
		},
	})

	populateServiceRecommendation(service, nil, os)

	assert.Assert(t, os.Len() == 0, "Should not populate one serviceScaleRecommendation object")
}

func TestPopulateAutoscalerMinMaxNotEqual(t *testing.T) {
	os := objectset.NewObjectSet()
	input := riov1.NewService("default", "test", riov1.Service{
		Spec: riov1.ServiceSpec{
			AutoscaleConfig: riov1.AutoscaleConfig{
				MinScale:    &[]int{0}[0],
				MaxScale:    &[]int{10}[0],
				Concurrency: &[]int{10}[0],
			},
			ServiceRevision: riov1.ServiceRevision{
				App:     "foo",
				Version: "v0",
			},
			PodConfig: riov1.PodConfig{
				Container: riov1.Container{
					Ports: []riov1.ContainerPort{
						{
							TargetPort: 8080,
						},
					},
				},
			},
		},
	})

	expected := constructors.NewPodAutoscaler(input.Namespace, input.Name, v1alpha1.PodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				ReferingLabel:         "kpa.autoscaling.knative.dev",
				MinScaleAnnotationKey: strconv.Itoa(*input.Spec.MinScale),
				MaxScaleAnnotationKey: strconv.Itoa(*input.Spec.MaxScale),
				ScrapeKey:             "envoy",
			},
			Labels: map[string]string{
				ConfigurationKey: input.Name,
				ServiceKey:       input.Name,
				RevisionKey:      "foo-v0",
				ContainerPortKey: "8080",
				AppKey:           "foo",
				VersionKey:       "v0",
			},
		},
		Spec: v1alpha1.PodAutoscalerSpec{
			ContainerConcurrency: servingv1beta1.RevisionContainerConcurrencyType(*input.Spec.AutoscaleConfig.Concurrency),
			ScaleTargetRef: corev1.ObjectReference{
				Kind:       "ServiceScaleRecommendation",
				APIVersion: autoscalev1.SchemeGroupVersion.String(),
				Name:       input.Name,
			},
			ProtocolType: networking.ProtocolHTTP1,
		},
	})

	populatePodAutoscaler(input, nil, os)

	test.AssertObjects(t, expected, os)
}

func TestPopulateAutoscalerMinMaxEqual(t *testing.T) {
	os := objectset.NewObjectSet()
	service := riov1.NewService("default", "test", riov1.Service{
		Spec: riov1.ServiceSpec{
			AutoscaleConfig: riov1.AutoscaleConfig{
				MinScale:    &[]int{10}[0],
				MaxScale:    &[]int{10}[0],
				Concurrency: &[]int{10}[0],
			},
			ServiceRevision: riov1.ServiceRevision{
				App:     "foo",
				Version: "v0",
			},
			PodConfig: riov1.PodConfig{
				Container: riov1.Container{
					Ports: []riov1.ContainerPort{
						{
							TargetPort: 8080,
						},
					},
				},
			},
		},
	})

	populatePodAutoscaler(service, nil, os)

	assert.Assert(t, os.Len() == 0, "Should not populate one PodAutoscaler object")
}
