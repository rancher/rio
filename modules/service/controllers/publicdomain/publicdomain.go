package publicdomain

import (
	"context"
	"sort"

	"github.com/rancher/mapper/slice"
	"k8s.io/apimachinery/pkg/api/errors"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
)

const (
	servicePublicdomain = "service-publicdomain"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		services: rContext.Rio.Rio().V1().Service(),
		routers:  rContext.Rio.Rio().V1().Router(),
	}

	rContext.Rio.Rio().V1().PublicDomain().OnChange(ctx, servicePublicdomain, h.sync)
	return nil
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
			deepcopy.Status.PublicDomains = append(deepcopy.Status.PublicDomains, pd.Spec.DomainName)
			sort.Strings(deepcopy.Status.PublicDomains)
		}
		if _, err := h.services.Update(deepcopy); err != nil {
			return pd, err
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
