package config

import (
	"context"
	"fmt"

	config2 "github.com/rancher/rio/pkg/config"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	serviceName = "gateway-proxy"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := &handler{
		key:             fmt.Sprintf("%s/%s", rContext.Namespace, config2.ConfigName),
		namespace:       rContext.Namespace,
		configMapClient: rContext.Core.Core().V1().ConfigMap(),
		serviceCache:    rContext.Core.Core().V1().Service().Cache(),
	}

	rContext.Core.Core().V1().ConfigMap().OnChange(ctx, "gloo-config", h.onChange)
	return nil
}

type handler struct {
	key             string
	namespace       string
	configMapClient corev1controller.ConfigMapClient
	serviceCache    corev1controller.ServiceCache
}

func (h *handler) onChange(key string, cm *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	if cm == nil || key != h.key {
		return nil, nil
	}

	config, err := config2.FromConfigMap(cm)
	if err != nil {
		return cm, err
	}

	if config.Gateway.ServiceNamespace == h.namespace &&
		config.Gateway.ServiceName == serviceName {
		return cm, nil
	}

	config.Gateway.ServiceNamespace = h.namespace
	config.Gateway.ServiceName = serviceName

	newCM, err := config2.SetConfig(cm, config)
	if err != nil {
		return cm, err
	}

	return h.configMapClient.Update(newCM)
}
