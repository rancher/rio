package serviceset

import (
	"context"
	"fmt"
	"sort"

	"github.com/rancher/rio/modules/service/controllers/service/populate/serviceports"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	services2 "github.com/rancher/rio/pkg/services"
	"github.com/rancher/rio/pkg/serviceset"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	v1 "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/objectset"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		namespace:      rContext.Namespace,
		apply:          rContext.Apply.WithSetID("serviceset"),
		apps:           rContext.Rio.Rio().V1().App(),
		coreservice:    rContext.Core.Core().V1().Service(),
		services:       rContext.Rio.Rio().V1().Service(),
		serviceCache:   rContext.Rio.Rio().V1().Service().Cache(),
		namespaceCache: rContext.Core.Core().V1().Namespace().Cache(),
	}
	rContext.Rio.Rio().V1().Service().OnChange(ctx, "serviceset-controller", h.onChange)
	return nil
}

type handler struct {
	namespace      string
	apply          apply.Apply
	coreservice    corev1controller.ServiceController
	services       riov1controller.ServiceController
	apps           riov1controller.AppController
	serviceCache   riov1controller.ServiceCache
	namespaceCache v1.NamespaceCache
}

func (h *handler) onChange(key string, service *riov1.Service) (*riov1.Service, error) {
	os := objectset.NewObjectSet()
	if service == nil {
		return service, nil
	}

	appName, _ := services2.AppAndVersion(service)

	ns, err := h.namespaceCache.Get(service.Namespace)
	if err != nil {
		if errors.IsNotFound(err) {
			return service, nil
		}
		return service, err
	}

	services, err := h.serviceCache.List(service.Namespace, labels.Everything())
	if err != nil {
		return service, err
	}

	serviceSet, err := serviceset.CollectionServices(services)
	if err != nil {
		return service, err
	}
	filteredServices := serviceSet[appName]
	if filteredServices == nil {
		return service, h.apply.WithSetID(appName).
			WithCacheTypes(h.apps).
			WithCacheTypes(h.coreservice).Apply(os)
	}

	svc := createService(ns.Name, appName, filteredServices.Revisions)
	os.Add(svc)

	// ServiceSet
	app := riov1.NewApp(service.Namespace, appName, riov1.App{
		Spec: riov1.AppSpec{
			Revisions: make([]riov1.Revision, 0),
		},
		Status: riov1.AppStatus{
			RevisionWeight: make(map[string]riov1.ServiceObservedWeight, 0),
		},
	})

	var totalweight int
	var serviceWeight []riov1.Revision
	for _, service := range filteredServices.Revisions {
		if service.DeletionTimestamp != nil {
			continue
		}
		_, version := services2.AppAndVersion(service)
		public := false
		for _, port := range service.Spec.Ports {
			if !port.InternalOnly {
				public = true
				break
			}
		}
		scale := service.Spec.Scale
		if service.Status.ObservedScale != nil {
			scale = service.Status.ObservedScale
		}

		scaleNum := 0
		if scale != nil {
			scaleNum = *scale
		}

		scaleStatus := service.Status.ScaleStatus
		weight := service.Spec.Weight

		// hack for daemonsets
		if scaleStatus == nil && service.Spec.Global {
			scaleStatus = &riov1.ScaleStatus{
				Available: scaleNum,
				Ready:     scaleNum,
			}
			weight = 100
		}

		serviceWeight = append(serviceWeight, riov1.Revision{
			Public:          public,
			Weight:          weight,
			ServiceName:     service.Name,
			Version:         version,
			Scale:           scaleNum,
			ScaleStatus:     scaleStatus,
			RolloutConfig:   service.Spec.RolloutConfig,
			DeploymentReady: IsReady(service.Status.DeploymentStatus),
		})
		totalweight += service.Spec.Weight
	}
	var added int
	for i, rev := range serviceWeight {
		if i == len(serviceWeight)-1 {
			rev.AdjustedWeight = 100 - added
		} else {
			if totalweight == 0 {
				rev.AdjustedWeight = int(1.0 / float64(len(serviceWeight)) * 100)
			} else {
				rev.AdjustedWeight = int(float64(rev.Weight) / float64(totalweight) * 100.0)
			}
			added += rev.AdjustedWeight
		}
		serviceWeight[i] = rev
	}
	sort.Slice(serviceWeight, func(i, j int) bool {
		return serviceWeight[i].Version < serviceWeight[j].Version
	})
	if len(serviceWeight) > 0 {
		app.Spec.Revisions = serviceWeight
		os.Add(app)
	}
	return service, h.apply.WithSetID(fmt.Sprintf("%s/%s", ns.Name, appName)).
		WithCacheTypes(h.apps).
		WithCacheTypes(h.coreservice).Apply(os)
}

func IsReady(status *appv1.DeploymentStatus) bool {
	if status == nil {
		return false
	}
	for _, con := range status.Conditions {
		if con.Type == appv1.DeploymentAvailable && con.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func createService(namespace, app string, serviceSet []*riov1.Service) *v12.Service {
	ports := portsForService(serviceSet)
	return constructors.NewService(namespace, app, v12.Service{
		Spec: v12.ServiceSpec{
			Ports: ports,
			Selector: map[string]string{
				"app": app,
			},
			Type: v12.ServiceTypeClusterIP,
		},
	})
}

func portsForService(serviceSet []*riov1.Service) (result []v12.ServicePort) {
	ports := map[struct {
		Port     int32
		Protocol v12.Protocol
	}]v12.ServicePort{}

	for _, rev := range serviceSet {
		for _, port := range serviceports.ServiceNamedPorts(rev) {
			ports[struct {
				Port     int32
				Protocol v12.Protocol
			}{
				Port:     port.Port,
				Protocol: port.Protocol,
			}] = port
		}
	}

	for _, port := range ports {
		result = append(result, port)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Port < result[j].Port
	})

	if len(result) == 0 {
		return []v12.ServicePort{
			{
				Name:       "default",
				Protocol:   v12.ProtocolTCP,
				TargetPort: intstr.FromInt(80),
				Port:       80,
			},
		}
	}
	return
}
