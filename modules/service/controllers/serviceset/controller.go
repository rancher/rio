package serviceset

import (
	"context"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/rancher/wrangler/pkg/kv"

	"github.com/rancher/rio/modules/service/controllers/service/populate/serviceports"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	v1 "github.com/rancher/rio/pkg/generated/controllers/core/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	services2 "github.com/rancher/rio/pkg/services"
	"github.com/rancher/rio/pkg/serviceset"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/objectset"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		namespace:      rContext.Namespace,
		apply:          rContext.Apply.WithSetID("serviceset"),
		apps:           rContext.Rio.Rio().V1().App(),
		services:       rContext.Rio.Rio().V1().Service(),
		serviceCache:   rContext.Rio.Rio().V1().Service().Cache(),
		namespaceCache: rContext.Core.Core().V1().Namespace().Cache(),
	}
	rContext.Rio.Rio().V1().Service().OnChange(ctx, "serviceset-controller", h.onChange)
	rContext.Rio.Rio().V1().Service().OnRemove(ctx, "serviceset-controller-remove", h.onChange)
	return nil
}

type handler struct {
	namespace      string
	apply          apply.Apply
	services       riov1controller.ServiceController
	apps           riov1controller.AppController
	serviceCache   riov1controller.ServiceCache
	namespaceCache v1.NamespaceCache
}

func (h *handler) onChange(key string, service *riov1.Service) (*riov1.Service, error) {
	if strings.Contains(key, "/") {
		namespace, _ := kv.Split(key, "/")
		h.services.Enqueue("", namespace)
		return nil, nil
	}

	ns, err := h.namespaceCache.Get(key)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	services, err := h.serviceCache.List(key, labels.Everything())
	if err != nil {
		return nil, err
	}

	serviceSet, err := serviceset.CollectionServices(services)

	os := objectset.NewObjectSet()
	for app, services := range serviceSet {
		svc := createService(ns.Name, app, services)
		os.Add(svc)
	}

	// ServiceSet
	for app, services := range serviceSet {
		app := riov1.NewApp(key, app, riov1.App{
			Spec: riov1.AppSpec{
				Revisions: make([]riov1.Revision, 0),
			},
			Status: riov1.AppStatus{
				RevisionWeight: make(map[string]riov1.ServiceObservedWeight, 0),
			},
		})

		var totalweight int
		var serviceWeight []riov1.Revision
		for _, service := range services.Revisions {
			_, version := services2.AppAndVersion(service)
			public := false
			for _, port := range service.Spec.Ports {
				if !port.InternalOnly {
					public = true
					break
				}
			}
			scale := service.Spec.Scale
			if scale == 0 {
				scale = 1
			}
			if service.Status.ObservedScale != nil && *service.Status.ObservedScale != 0 {
				scale = *service.Status.ObservedScale
			}
			serviceWeight = append(serviceWeight, riov1.Revision{
				Public:          public,
				Weight:          service.Spec.Weight,
				ServiceName:     service.Name,
				Version:         version,
				Scale:           scale,
				ScaleStatus:     service.Status.ScaleStatus,
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
		app.Spec.Revisions = serviceWeight
		os.Add(app)
	}

	return service, h.apply.WithOwner(ns).WithCacheTypes(h.apps).Apply(os)
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

func createService(namespace, app string, serviceSet *serviceset.ServiceSet) *v12.Service {
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

func portsForService(serviceSet *serviceset.ServiceSet) (result []v12.ServicePort) {
	ports := map[struct {
		Port     int32
		Protocol v12.Protocol
	}]v12.ServicePort{}

	for _, rev := range serviceSet.Revisions {
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
