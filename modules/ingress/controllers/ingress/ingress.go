package ingress

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rancher/rio/modules/gloo/pkg/vsfactory"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/indexes"
	"github.com/rancher/rio/pkg/serviceports"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/name"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
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
		"ingress-services",
		h.generateFromService,
		nil)

	corev1controller.RegisterServiceGeneratingHandler(ctx,
		rContext.Core.Core().V1().Service(),
		rContext.Apply.
			WithCacheTypes(
				rContext.K8sNetworking.Extensions().V1beta1().Ingress(),
				rContext.Core.Core().V1().Secret()),
		"",
		"ingress-app",
		h.generateFromApp,
		nil)
	return nil
}

type handler struct {
	namespace    string
	vsFactory    *vsfactory.VirtualServiceFactory
	secrets      corev1controller.SecretController
	serviceCache riov1controller.ServiceCache
}

func (h handler) generateFromService(obj *riov1.Service, status riov1.ServiceStatus) ([]runtime.Object, riov1.ServiceStatus, error) {
	if obj.Spec.Template {
		return nil, status, nil
	}

	app, version := services.AppAndVersion(obj)

	servicePort := servicePort(obj)
	if servicePort == 0 {
		return nil, status, nil
	}

	hostnames := hostnames(obj.Status.Endpoints)

	objects, err := h.IngressAndSecrets(obj.Namespace, app, version, servicePort, hostnames)
	return objects, status, err
}

func (h handler) generateFromApp(svc *corev1.Service, status corev1.ServiceStatus) ([]runtime.Object, corev1.ServiceStatus, error) {
	app := svc.Labels["rio.cattle.io/app"]
	if app == "" {
		return nil, status, nil
	}

	svcs, err := h.serviceCache.GetByIndex(indexes.ServiceByApp, fmt.Sprintf("%s/%s", svc.Namespace, app))
	if err != nil {
		return nil, status, err
	}

	if len(svcs) == 0 {
		return nil, status, nil
	}
	servicePort := servicePort(svcs[0])
	if servicePort == 0 {
		return nil, status, nil
	}

	hostnames := hostnames(svcs[0].Status.AppEndpoints)
	objects, err := h.IngressAndSecrets(svc.Namespace, app, "", servicePort, hostnames)
	return objects, status, nil
}

func (h handler) IngressAndSecrets(namespace, app, version string, servicePort int32, hostNames []string) ([]runtime.Object, error) {
	var result []runtime.Object

	tlss, err := h.vsFactory.FindTLS(namespace, app, version, hostNames)
	if err != nil {
		return nil, err
	}
	svcName := app
	if version != "" {
		svcName = name.SafeConcatName(app, version)
	}

	ingress := constructors.NewIngress(namespace, svcName, v1beta1.Ingress{})

	for _, hostname := range hostNames {
		ingress.Spec.Rules = append(ingress.Spec.Rules, v1beta1.IngressRule{
			Host: hostname,
			IngressRuleValue: v1beta1.IngressRuleValue{
				HTTP: &v1beta1.HTTPIngressRuleValue{
					Paths: []v1beta1.HTTPIngressPath{
						{
							Backend: v1beta1.IngressBackend{
								ServiceName: svcName,
								ServicePort: intstr.FromInt(int(servicePort)),
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
			secret := constructors.NewSecret(namespace, tlss[hostname], corev1.Secret{
				Data: existingSecret.Data,
			})
			if _, err := h.secrets.Create(secret); err != nil && !errors.IsAlreadyExists(err) {
				return nil, err
			}
		}
	}
	result = append(result, ingress)
	return result, nil
}

func servicePort(svc *riov1.Service) int32 {
	var servicePort int32
	for _, port := range serviceports.ContainerPorts(svc) {
		if port.IsExposed() && port.IsHTTP() {
			servicePort = port.Port
			continue
		}
	}
	return servicePort
}

func hostnames(endpoints []string) []string {
	var hostnames []string
	seen := map[string]bool{}
	for _, endpoint := range endpoints {
		u, err := url.Parse(endpoint)
		if err != nil {
			continue
		}
		if seen[u.Host] {
			continue
		}
		seen[u.Host] = true

		hostnames = append(hostnames, u.Host)
	}
	return hostnames
}
