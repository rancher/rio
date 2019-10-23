package router

import (
	"context"
	"fmt"

	config "github.com/rancher/rio/pkg/config"

	corev1 "k8s.io/api/core/v1"

	"github.com/rancher/wrangler/pkg/generic"

	v1 "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"

	"github.com/rancher/rio/modules/service/controllers/router/populate"
	"github.com/rancher/rio/modules/service/pkg/endpoints"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := &handler{
		configKey:        fmt.Sprintf("%s/%s", rContext.Namespace, config.ConfigName),
		systemNamespace:  rContext.Namespace,
		configMapCache:   rContext.Core.Core().V1().ConfigMap().Cache(),
		routerController: rContext.Rio.Rio().V1().Router(),
		resolver: endpoints.NewResolver(ctx, rContext.Namespace,
			rContext.Rio.Rio().V1().Router(),
			rContext.Rio.Rio().V1().Service().Cache(),
			rContext.Admin.Admin().V1().ClusterDomain(),
			rContext.Admin.Admin().V1().PublicDomain()),
	}

	rContext.Core.Core().V1().ConfigMap().OnChange(ctx, "router-config", h.onConfigMap)

	riov1controller.RegisterRouterGeneratingHandler(ctx,
		rContext.Rio.Rio().V1().Router(),
		rContext.Apply.WithCacheTypes(rContext.Core.Core().V1().Service(),
			rContext.Core.Core().V1().Endpoints()),
		"RouterDeployed",
		"router",
		h.generate,
		nil)

	return nil
}

type handler struct {
	configKey        string
	systemNamespace  string
	gatewayName      string
	gatewayNamespace string
	configMapCache   v1.ConfigMapCache
	routerController riov1controller.RouterController
	resolver         *endpoints.Resolver
}

func (h *handler) onConfigMap(key string, cm *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	if cm == nil || key != h.configKey {
		return cm, nil
	}

	config, err := config.FromConfigMap(cm)
	if err != nil {
		return cm, err
	}

	if h.gatewayName == config.Gateway.ServiceName &&
		h.gatewayNamespace == config.Gateway.ServiceNamespace {
		return cm, nil
	}

	h.gatewayName = config.Gateway.ServiceName
	h.gatewayNamespace = config.Gateway.ServiceNamespace
	h.routerController.Enqueue("*", "*")

	return cm, nil
}

func (h *handler) generate(obj *riov1.Router, status riov1.RouterStatus) ([]runtime.Object, riov1.RouterStatus, error) {
	if h.gatewayNamespace == "" || h.gatewayName == "" {
		return nil, status, generic.ErrSkip
	}

	os := objectset.NewObjectSet()
	if err := populate.ServiceForRouter(obj, h.gatewayNamespace, h.gatewayName, os); err != nil {
		return nil, status, err
	}

	endpoints, err := h.resolver.RouterEndpoints(obj)
	if err != nil {
		return nil, status, err
	}

	status.Endpoints = endpoints
	return os.All(), status, err
}
