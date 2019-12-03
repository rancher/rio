package ingress

import (
	"context"
	"net/url"

	"github.com/rancher/rio/modules/gloo/pkg/vsfactory"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/serviceports"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/name"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		namespace: rContext.Namespace,
		secrets:   rContext.Core.Core().V1().Secret(),
		vsFactory: vsfactory.New(rContext),
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
	return nil
}

type handler struct {
	namespace string
	vsFactory *vsfactory.VirtualServiceFactory
	secrets   corev1controller.SecretController
}

func (h handler) generateFromService(obj *riov1.Service, status riov1.ServiceStatus) ([]runtime.Object, riov1.ServiceStatus, error) {
	if obj.Spec.Template {
		return nil, status, nil
	}

	app, version := services.AppAndVersion(obj)

	var servicePort int32
	for _, port := range serviceports.ContainerPorts(obj) {
		if port.IsExposed() && port.IsHTTP() {
			servicePort = port.Port
			continue
		}
	}
	if servicePort == 0 {
		return nil, status, nil
	}

	var hostnames []string
	seen := map[string]bool{}
	for _, endpoint := range obj.Status.Endpoints {
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

	tlss, err := h.vsFactory.FindTLS(obj.Namespace, app, version, hostnames)
	if err != nil {
		return nil, status, err
	}

	var result []runtime.Object

	ingress := constructors.NewIngress(obj.Namespace, name.SafeConcatName(app, version), v1beta1.Ingress{})

	for _, hostname := range hostnames {
		ingress.Spec.Rules = append(ingress.Spec.Rules, v1beta1.IngressRule{
			Host: hostname,
			IngressRuleValue: v1beta1.IngressRuleValue{
				HTTP: &v1beta1.HTTPIngressRuleValue{
					Paths: []v1beta1.HTTPIngressPath{
						{
							Backend: v1beta1.IngressBackend{
								ServiceName: name.SafeConcatName(app, version),
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
				return nil, status, err
			}
			secret := constructors.NewSecret(obj.Namespace, tlss[hostname], corev1.Secret{
				Data: existingSecret.Data,
			})

			result = append(result, secret)
		}
	}
	result = append(result, ingress)

	return result, status, nil
}
