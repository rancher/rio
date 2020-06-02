package service

import (
	"context"

	"github.com/rancher/rio/modules/gloo/pkg/vsfactory"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	rioadminv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := &handler{
		systemNamespace:    rContext.Namespace,
		clusterDomainCache: rContext.Admin.Admin().V1().ClusterDomain().Cache(),
		vsFactory:          vsfactory.New(rContext),
	}

	riov1controller.RegisterServiceGeneratingHandler(ctx,
		rContext.Rio.Rio().V1().Service(),
		rContext.Apply.WithCacheTypes(
			rContext.Gateway.Gateway().V1().VirtualService(),
		),
		"",
		"gloo-service",
		h.generate,
		nil)

	return nil
}

type handler struct {
	systemNamespace    string
	clusterDomainCache rioadminv1controller.ClusterDomainCache
	vsFactory          *vsfactory.VirtualServiceFactory
}

func (h *handler) generate(obj *riov1.Service, status riov1.ServiceStatus) ([]runtime.Object, riov1.ServiceStatus, error) {
	if obj.Spec.Template {
		return nil, status, nil
	}
	vss, err := h.vsFactory.ForRevision(obj)
	if err != nil {
		return nil, status, err
	}
	var result []runtime.Object
	for _, vs := range vss {
		result = append(result, vs)
	}
	return result, status, nil
}
