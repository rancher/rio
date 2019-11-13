package app

import (
	"context"
	"fmt"
	"sort"

	"github.com/deislabs/smi-sdk-go/pkg/apis/split/v1alpha1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/indexes"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/generic"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := &handler{
		serviceCache: rContext.Rio.Rio().V1().Service().Cache(),
	}

	corev1controller.RegisterServiceGeneratingHandler(ctx,
		rContext.Core.Core().V1().Service(),
		rContext.Apply.
			WithCacheTypes(rContext.SMI.Split().V1alpha1().TrafficSplit()),
		"",
		"smi",
		h.generate,
		nil)

	return nil
}

type handler struct {
	serviceCache riov1controller.ServiceCache
}

func (h *handler) generate(service *corev1.Service, status corev1.ServiceStatus) ([]runtime.Object, corev1.ServiceStatus, error) {
	app := service.Labels["rio.cattle.io/app"]
	if app == "" {
		return nil, status, generic.ErrSkip
	}

	svcs, err := h.serviceCache.GetByIndex(indexes.ServiceByApp, fmt.Sprintf("%s/%s", service.Namespace, app))
	if err != nil {
		return nil, status, err
	}

	if len(svcs) == 0 {
		return nil, status, nil
	}

	ts := h.getSplit(service.Namespace, svcs)
	if ts == nil {
		return nil, status, nil
	}
	return []runtime.Object{
		ts,
	}, status, nil
}

func (h *handler) getSplit(namespace string, svcs []*riov1.Service) *v1alpha1.TrafficSplit {
	var (
		backends []v1alpha1.TrafficSplitBackend
		service  string
	)

	for _, svc := range svcs {
		if svc.Status.ComputedWeight == nil || *svc.Status.ComputedWeight <= 0 {
			continue
		}

		app, version := services.AppAndVersion(svc)
		service = app
		quantity := resource.NewQuantity(int64(*svc.Status.ComputedWeight), resource.BinarySI)
		backends = append(backends, v1alpha1.TrafficSplitBackend{
			Service: fmt.Sprintf("%s-%s", app, version),
			Weight:  *quantity,
		})
	}

	sort.Slice(backends, func(i, j int) bool {
		return backends[i].Service < backends[j].Service
	})

	if service == "" {
		return nil
	}

	return &v1alpha1.TrafficSplit{
		ObjectMeta: metav1.ObjectMeta{
			Name:      service,
			Namespace: namespace,
		},
		Spec: v1alpha1.TrafficSplitSpec{
			Service:  service,
			Backends: backends,
		},
	}
}
