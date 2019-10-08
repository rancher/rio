package app

import (
	"context"
	"fmt"
	"sort"

	"github.com/rancher/rio/modules/service/controllers/service/populate/serviceports"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/indexes"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/relatedresource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Register(ctx context.Context, rContext *types.Context) error {
	appSelector, err := labels.Parse("rio.cattle.io/app")
	if err != nil {
		return err
	}

	h := &handler{
		serviceCache: rContext.Rio.Rio().V1().Service().Cache(),
		services:     rContext.Core.Core().V1().Service(),
		appSelector:  appSelector,
	}

	rContext.Rio.Rio().V1().Service().OnChange(ctx, "app", h.onRioServiceChange)
	rContext.Core.Core().V1().Service().OnChange(ctx, "app", h.onServiceChange)

	relatedresource.Watch(ctx, "app",
		resolveAppService,
		rContext.Core.Core().V1().Service(),
		rContext.Rio.Rio().V1().Service())

	return nil
}

func resolveAppService(_, _ string, obj runtime.Object) ([]relatedresource.Key, error) {
	svc, ok := obj.(*riov1.Service)
	if !ok {
		return nil, nil
	}

	app, _ := services.AppAndVersion(svc)
	return []relatedresource.Key{
		{
			Namespace: svc.Namespace,
			Name:      app,
		},
	}, nil
}

type handler struct {
	serviceCache riov1controller.ServiceCache
	services     corev1controller.ServiceController
	appSelector  labels.Selector
}

func (h *handler) onRioServiceChange(key string, svc *riov1.Service) (*riov1.Service, error) {
	if svc == nil {
		return nil, nil
	}

	appName, _ := services.AppAndVersion(svc)
	revisions, err := h.serviceCache.GetByIndex(indexes.ServiceByApp, fmt.Sprintf("%s/%s", svc.Namespace, appName))
	if err != nil || len(revisions) == 0 {
		return svc, err
	}

	_, err = h.services.Cache().Get(svc.Namespace, appName)
	if err == nil {
		// Already Exists
		return svc, nil
	}

	appService := newService(svc.Namespace, appName, revisions)

	_, err = h.services.Create(appService)
	if errors.IsAlreadyExists(err) {
		// Already Exists
		return svc, nil
	}

	return svc, err
}

func (h *handler) onServiceChange(key string, svc *corev1.Service) (*corev1.Service, error) {
	if svc == nil {
		return nil, nil
	}

	appName := svc.Labels["rio.cattle.io/app"]
	if appName == "" {
		return svc, nil
	}

	values, err := h.serviceCache.GetByIndex(indexes.ServiceByApp, fmt.Sprintf("%s/%s", svc.Namespace, appName))
	if err != nil {
		return svc, err
	}

	if len(values) == 0 {
		return svc, h.services.Delete(svc.Namespace, svc.Name, nil)
	}

	return svc, nil
}

func newService(namespace, app string, serviceSet []*riov1.Service) *corev1.Service {
	ports := portsForService(serviceSet)
	return constructors.NewService(namespace, app, corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"rio.cattle.io/app": app,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: ports,
			Selector: map[string]string{
				"app": app,
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	})
}

func portsForService(serviceSet []*riov1.Service) (result []corev1.ServicePort) {
	ports := map[string]corev1.ServicePort{}

	for _, rev := range serviceSet {
		for _, port := range serviceports.ServiceNamedPorts(rev) {
			ports[port.Name] = port
		}
	}

	for _, port := range ports {
		result = append(result, port)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	if len(result) == 0 {
		return []corev1.ServicePort{
			{
				Name:       "default",
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(80),
				Port:       80,
			},
		}
	}

	return
}
