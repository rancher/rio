package publicdomain

import (
	"context"
	"sort"

	"github.com/rancher/mapper/slice"
	"github.com/rancher/rio/modules/service/controllers/publicdomain/populate"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	servicePublicdomain = "service-publicdomain"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "stack-public-domain", rContext.Rio.Rio().V1().PublicDomain())
	c.Apply = c.Apply.WithCacheTypes(rContext.Extensions.Extensions().V1beta1().Ingress())

	p := populator{
		systemNamespace: rContext.Namespace,
	}

	c.Populator = p.populate

	h := handler{
		services: rContext.Rio.Rio().V1().Service(),
		routers:  rContext.Rio.Rio().V1().Router(),
	}

	rContext.Rio.Rio().V1().PublicDomain().OnChange(ctx, servicePublicdomain, h.sync)
	return nil
}

type populator struct {
	systemNamespace string
}

func (p populator) populate(obj runtime.Object, namespace *corev1.Namespace, os *objectset.ObjectSet) error {
	return populate.Ingress(p.systemNamespace, obj.(*riov1.PublicDomain), os)
}

type handler struct {
	services riov1controller.ServiceController
	routers  riov1controller.RouterController
}

func (h handler) sync(key string, pd *riov1.PublicDomain) (*riov1.PublicDomain, error) {
	if pd == nil {
		return nil, nil
	}

	svc, err := h.services.Cache().Get(pd.Namespace, pd.Spec.TargetServiceName)
	if err != nil && !errors.IsNotFound(err) {
		return pd, err
	}

	if err == nil {
		deepcopy := svc.DeepCopy()
		if !slice.ContainsString(deepcopy.Status.PublicDomains, pd.Spec.DomainName) {
			if !slice.ContainsString(deepcopy.Status.PublicDomains, pd.Spec.DomainName) {
				deepcopy.Status.PublicDomains = append(deepcopy.Status.PublicDomains, pd.Spec.DomainName)
				sort.Strings(deepcopy.Status.PublicDomains)
				if _, err := h.services.Update(deepcopy); err != nil {
					return pd, err
				}
			}
		}
		return pd, nil
	}

	router, err := h.routers.Cache().Get(pd.Namespace, pd.Spec.TargetServiceName)
	if err != nil && !errors.IsNotFound(err) {
		return pd, err
	}

	if err == nil {
		deepcopy := router.DeepCopy()
		if !slice.ContainsString(deepcopy.Status.PublicDomains, pd.Spec.DomainName) {
			deepcopy.Status.PublicDomains = append(deepcopy.Status.PublicDomains, pd.Spec.DomainName)
			sort.Strings(deepcopy.Status.PublicDomains)
		}
		if _, err := h.routers.Update(deepcopy); err != nil {
			return pd, err
		}
		return pd, nil
	}

	return pd, nil
}
