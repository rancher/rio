package config

import (
	"context"
	"fmt"

	config2 "github.com/rancher/rio/pkg/config"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	ingressName = "rio-cluster-ingress"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := &handler{
		key:             fmt.Sprintf("%s/%s", rContext.Namespace, config2.ConfigName),
		namespace:       rContext.Namespace,
		configMapClient: rContext.Core.Core().V1().ConfigMap(),
		serviceCache:    rContext.Core.Core().V1().Service().Cache(),
	}

	ingress := constructors.NewIngress(rContext.Namespace, ingressName, v1beta1.Ingress{})
	ingress.Spec.Rules = append(ingress.Spec.Rules, v1beta1.IngressRule{
		Host: "fake.gateway.rio.io",
		IngressRuleValue: v1beta1.IngressRuleValue{
			HTTP: &v1beta1.HTTPIngressRuleValue{
				Paths: []v1beta1.HTTPIngressPath{
					{
						Backend: v1beta1.IngressBackend{
							ServiceName: constants.AuthWebhookServiceName,
							ServicePort: intstr.FromInt(80),
						},
					},
				},
			},
		},
	})
	if _, err := rContext.K8sNetworking.Extensions().V1beta1().Ingress().Get(rContext.Namespace, ingressName, metav1.GetOptions{}); errors.IsNotFound(err) {
		if _, err := rContext.K8sNetworking.Extensions().V1beta1().Ingress().Create(ingress); err != nil {
			return err
		}
	}

	rContext.Core.Core().V1().ConfigMap().OnChange(ctx, "ingress-config", h.onChange)
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

	if config.Gateway.IngressName == ingressName &&
		config.Gateway.IngressNamespace == h.namespace {
		return cm, nil
	}

	config.Gateway.IngressNamespace = h.namespace
	config.Gateway.IngressName = ingressName

	newCM, err := config2.SetConfig(cm, config)
	if err != nil {
		return cm, err
	}

	return h.configMapClient.Update(newCM)
}
