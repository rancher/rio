package serviceset

import (
	"context"
	"sort"
	"strings"

	"github.com/rancher/rio/modules/service/controllers/service/populate/serviceports"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	v1 "github.com/rancher/rio/pkg/generated/controllers/core/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/serviceset"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/objectset"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		apply:          rContext.Apply.WithSetID("serviceset"),
		services:       rContext.Rio.Rio().V1().Service(),
		serviceCache:   rContext.Rio.Rio().V1().Service().Cache(),
		namespaceCache: rContext.Core.Core().V1().Namespace().Cache(),
	}
	rContext.Rio.Rio().V1().Service().OnChange(ctx, "serviceset-controller", h.onChange)
	return nil
}

type handler struct {
	apply          apply.Apply
	services       riov1controller.ServiceController
	serviceCache   riov1controller.ServiceCache
	namespaceCache v1.NamespaceCache
}

func (h *handler) onChange(key string, service *riov1.Service) (*riov1.Service, error) {
	if strings.Contains(key, "/") {
		h.services.Enqueue("", service.Namespace)
		return nil, nil
	}

	ns, err := h.namespaceCache.Get(key)
	if err != nil {
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

	return service, h.apply.WithOwner(ns).Apply(os)
}

func createService(namespace, app string, serviceSet *serviceset.ServiceSet) *v12.Service {
	return constructors.NewService(namespace, app, v12.Service{
		Spec: v12.ServiceSpec{
			Ports: portsForService(serviceSet),
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

	return
}
