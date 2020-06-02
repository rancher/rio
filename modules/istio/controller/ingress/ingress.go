package ingress

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rancher/rio/modules/istio/controller/pkg"

	"github.com/rancher/rio/modules/gloo/pkg/vsfactory"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/config"
	"github.com/rancher/rio/pkg/constructors"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/indexes"
	"github.com/rancher/rio/pkg/serviceports"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/generic"
	"github.com/rancher/wrangler/pkg/name"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		namespace:    rContext.Namespace,
		secrets:      rContext.Core.Core().V1().Secret(),
		vsFactory:    vsfactory.New(rContext),
		serviceCache: rContext.Rio.Rio().V1().Service().Cache(),
	}

	riov1controller.RegisterServiceGeneratingHandler(ctx,
		rContext.Rio.Rio().V1().Service(),
		rContext.Apply.
			WithCacheTypes(
				rContext.K8sNetworking.Extensions().V1beta1().Ingress(),
				rContext.Core.Core().V1().Secret()),
		"",
		"ingress-istio-services",
		h.generateFromService,
		&generic.GeneratingHandlerOptions{
			AllowCrossNamespace: true,
		})

	riov1controller.RegisterRouterGeneratingHandler(ctx,
		rContext.Rio.Rio().V1().Router(),
		rContext.Apply.WithCacheTypes(
			rContext.K8sNetworking.Extensions().V1beta1().Ingress(),
			rContext.Core.Core().V1().Secret()),
		"",
		"ingress-istio-router",
		h.generateFromRouter,
		&generic.GeneratingHandlerOptions{
			AllowCrossNamespace: true,
		})

	corev1controller.RegisterServiceGeneratingHandler(ctx,
		rContext.Core.Core().V1().Service(),
		rContext.Apply.
			WithCacheTypes(
				rContext.K8sNetworking.Extensions().V1beta1().Ingress(),
				rContext.Core.Core().V1().Secret()),
		"",
		"ingress-istio-app-services",
		h.generateAppFromService,
		&generic.GeneratingHandlerOptions{
			AllowCrossNamespace: true,
		})
	return nil
}

type handler struct {
	namespace    string
	vsFactory    *vsfactory.VirtualServiceFactory
	secrets      corev1controller.SecretController
	serviceCache riov1controller.ServiceCache
}

func (h handler) generateAppFromService(obj *corev1.Service, status corev1.ServiceStatus) ([]runtime.Object, corev1.ServiceStatus, error) {
	app := obj.Labels["rio.cattle.io/app"]
	if app == "" {
		return nil, status, nil
	}

	svcs, err := h.serviceCache.GetByIndex(indexes.ServiceByApp, fmt.Sprintf("%s/%s", obj.Namespace, app))
	if err != nil {
		return nil, status, err
	}

	if len(svcs) == 0 {
		return nil, status, nil
	}

	service := svcs[0]
	result, err := h.generateIngressAndSecret(service, service.Status.AppEndpoints, true)
	return result, status, err
}

func (h handler) generateFromService(obj *riov1.Service, status riov1.ServiceStatus) ([]runtime.Object, riov1.ServiceStatus, error) {
	if obj.Spec.Template {
		return nil, status, nil
	}

	result, err := h.generateIngressAndSecret(obj, obj.Status.Endpoints, false)
	return result, status, err
}

func (h handler) generateFromRouter(router *riov1.Router, status riov1.RouterStatus) ([]runtime.Object, riov1.RouterStatus, error) {
	dms, err := pkg.Domains(router)
	if err != nil {
		return nil, status, err
	}

	tlss, err := h.vsFactory.FindTLS(router.Namespace, router.Name, "", dms)
	if err != nil {
		return nil, status, err
	}
	var result []runtime.Object

	ingress := constructors.NewIngress(config.ConfigController.Gateway.ServiceNamespace, router.Name, v1beta1.Ingress{})

	for _, hostname := range dms {
		ingress.Spec.Rules = append(ingress.Spec.Rules, v1beta1.IngressRule{
			Host: hostname,
			IngressRuleValue: v1beta1.IngressRuleValue{
				HTTP: &v1beta1.HTTPIngressRuleValue{
					Paths: []v1beta1.HTTPIngressPath{
						{
							Backend: v1beta1.IngressBackend{
								ServiceName: config.ConfigController.Gateway.ServiceName,
								ServicePort: intstr.FromInt(80),
							},
						},
					},
				},
			},
		})
		ingress.Spec.TLS = append(ingress.Spec.TLS, v1beta1.IngressTLS{
			Hosts:      []string{hostname},
			SecretName: tlss[hostname],
		})

		if tlss[hostname] != "" {
			existingSecret, err := h.secrets.Cache().Get(h.namespace, tlss[hostname])
			if err != nil {
				return nil, status, err
			}
			secret := constructors.NewSecret(config.ConfigController.Gateway.ServiceNamespace, tlss[hostname], corev1.Secret{
				Data: existingSecret.Data,
			})

			result = append(result, secret)
		}
	}
	result = append(result, ingress)
	return result, status, nil

}

func (h handler) generateIngressAndSecret(obj *riov1.Service, endpoints []string, isApp bool) ([]runtime.Object, error) {
	app, version := services.AppAndVersion(obj)

	var servicePort int32
	for _, port := range serviceports.ContainerPorts(obj) {
		if port.IsExposed() && port.IsHTTP() {
			servicePort = port.Port
			continue
		}
	}
	if servicePort == 0 {
		return nil, nil
	}

	var hostnames []string
	seen := map[string]bool{}
	for _, endpoint := range endpoints {
		u, err := url.Parse(endpoint)
		if err != nil {
			continue
		}
		if seen[u.Hostname()] {
			continue
		}
		seen[u.Hostname()] = true

		hostnames = append(hostnames, u.Hostname())
	}

	tlss, err := h.vsFactory.FindTLS(obj.Namespace, app, version, hostnames)
	if err != nil {
		return nil, err
	}

	var result []runtime.Object

	ingressName := name.SafeConcatName(app, version)
	if isApp {
		ingressName = app
	}
	ingress := constructors.NewIngress(config.ConfigController.Gateway.ServiceNamespace, ingressName, v1beta1.Ingress{})

	for _, hostname := range hostnames {
		ingress.Spec.Rules = append(ingress.Spec.Rules, v1beta1.IngressRule{
			Host: hostname,
			IngressRuleValue: v1beta1.IngressRuleValue{
				HTTP: &v1beta1.HTTPIngressRuleValue{
					Paths: []v1beta1.HTTPIngressPath{
						{
							Backend: v1beta1.IngressBackend{
								ServiceName: config.ConfigController.Gateway.ServiceName,
								ServicePort: intstr.FromInt(80),
							},
						},
					},
				},
			},
		})
		ingress.Spec.TLS = append(ingress.Spec.TLS, v1beta1.IngressTLS{
			Hosts:      []string{hostname},
			SecretName: tlss[hostname],
		})

		if tlss[hostname] != "" {
			existingSecret, err := h.secrets.Cache().Get(h.namespace, tlss[hostname])
			if err != nil {
				return nil, err
			}
			secret := constructors.NewSecret(config.ConfigController.Gateway.ServiceNamespace, tlss[hostname], corev1.Secret{
				Data: existingSecret.Data,
			})

			result = append(result, secret)
		}
	}
	result = append(result, ingress)
	return result, nil
}
