package gateway

import (
	"context"
	"fmt"
	"strings"

	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/config"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	adminv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	istiov1alpha3controller "github.com/rancher/rio/pkg/generated/controllers/networking.istio.io/v1alpha3"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/name"
	"github.com/rancher/wrangler/pkg/relatedresource"
	networkingv1alpha3 "istio.io/api/networking/v1alpha3"
	istiov1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		systemNamespace: rContext.Namespace,
		secrets:         rContext.Core.Core().V1().Secret().Cache(),
		configmaps:      rContext.Core.Core().V1().ConfigMap().Cache(),
		gateways:        rContext.Istio.Networking().V1alpha3().Gateway(),
		clusterdomains:  rContext.Admin.Admin().V1().ClusterDomain().Cache(),
		publicdomains:   rContext.Admin.Admin().V1().PublicDomain().Cache(),
	}

	relatedresource.Watch(ctx, "gateway-enqueue", h.resolver,
		enqueuer{Context: rContext},
		rContext.Core.Core().V1().Secret(),
		rContext.Core.Core().V1().ConfigMap(),
		rContext.Admin.Admin().V1().PublicDomain(),
		rContext.Admin.Admin().V1().ClusterDomain())

	adminv1controller.RegisterRioInfoGeneratingHandler(ctx,
		rContext.Admin.Admin().V1().RioInfo(),
		rContext.Apply.
			WithCacheTypes(rContext.Istio.Networking().V1alpha3().Gateway(),
				rContext.Core.Core().V1().Secret()),
		"",
		"istio-app",
		h.generate,
		nil)
	return nil
}

type enqueuer struct {
	Context *types.Context
}

func (e enqueuer) Enqueue(namespace, name string) {
	e.Context.Admin.Admin().V1().RioInfo().Enqueue(name)
}

func (h handler) resolver(namespace, name string, obj runtime.Object) ([]relatedresource.Key, error) {
	switch obj.(type) {
	case *v1.ConfigMap:
		if name == "rio" && namespace == h.systemNamespace {
			return []relatedresource.Key{
				{
					Namespace: h.systemNamespace,
					Name:      "rio",
				},
			}, nil
		}
	case *v1.Secret:
		if strings.HasSuffix(name, "tls") {
			return []relatedresource.Key{
				{
					Namespace: h.systemNamespace,
					Name:      "rio",
				},
			}, nil
		}
	default:
		return []relatedresource.Key{
			{
				Namespace: h.systemNamespace,
				Name:      "rio",
			},
		}, nil
	}
	return nil, nil
}

type handler struct {
	systemNamespace string
	secrets         corev1controller.SecretCache
	configmaps      corev1controller.ConfigMapCache
	gateways        istiov1alpha3controller.GatewayController
	clusterdomains  adminv1controller.ClusterDomainCache
	publicdomains   adminv1controller.PublicDomainCache
}

func (h handler) generate(obj *adminv1.RioInfo, status adminv1.RioInfoStatus) ([]runtime.Object, adminv1.RioInfoStatus, error) {
	if obj == nil || obj.DeletionTimestamp != nil {
		return nil, status, nil
	}

	if obj.Name != "rio" {
		return nil, status, nil
	}

	cm, err := h.configmaps.Get(h.systemNamespace, "rio")
	if err != nil {
		return nil, status, err
	}

	conf, err := config.FromConfigMap(cm)
	if err != nil {
		return nil, status, err
	}

	// set default labels selector, or can be overridden by feature configuration
	labelSelectors := map[string]string{
		"app": "istio-ingressgateway",
	}
	if conf.Features["istio"].Options != nil {
		parts := strings.Split(conf.Features["istio"].Options["labelSelectors"], "=")
		if len(parts) > 1 {
			labelSelectors[parts[0]] = parts[1]
		}
	}

	var result []runtime.Object

	var servers []*networkingv1alpha3.Server
	clusterdomains, err := h.clusterdomains.List(labels.Everything())
	if err != nil {
		return nil, status, err
	}
	publicdomains, err := h.publicdomains.List(labels.Everything())
	if err != nil {
		return nil, status, err
	}
	for _, cd := range clusterdomains {
		secretName := name.SafeConcatName(cd.Name, "tls")
		if cd.Spec.SecretName != "" {
			secretName = cd.Spec.SecretName
		}
		secret := h.copySecret(h.systemNamespace, secretName)
		if secret != nil {
			result = append(result, secret)
		}
		servers = append(servers,
			&networkingv1alpha3.Server{
				Port: &networkingv1alpha3.Port{
					Name:     "http",
					Protocol: "HTTP",
					Number:   80,
				},
				Hosts: []string{fmt.Sprintf("*.%s", cd.Name)},
			},
			&networkingv1alpha3.Server{
				Port: &networkingv1alpha3.Port{
					Name:     "https",
					Protocol: "HTTPS",
					Number:   443,
				},
				Hosts: []string{fmt.Sprintf("*.%s", cd.Name)},
				Tls: &networkingv1alpha3.ServerTLSSettings{
					Mode:           networkingv1alpha3.ServerTLSSettings_SIMPLE,
					CredentialName: secretName,
				},
			})
	}
	for _, cd := range publicdomains {
		secretName := name.SafeConcatName(cd.Name, "tls")
		if cd.Spec.SecretName != "" {
			secretName = cd.Spec.SecretName
		}
		secret := h.copySecret(h.systemNamespace, secretName)
		if secret != nil {
			result = append(result, secret)
		}
		servers = append(servers,
			&networkingv1alpha3.Server{
				Port: &networkingv1alpha3.Port{
					Name:     cd.Name + "-http",
					Protocol: "HTTP",
					Number:   80,
				},
				Hosts: []string{cd.Name},
			},
			&networkingv1alpha3.Server{
				Port: &networkingv1alpha3.Port{
					Name:     cd.Name + "-https",
					Protocol: "HTTPS",
					Number:   443,
				},
				Hosts: []string{cd.Name},
				Tls: &networkingv1alpha3.ServerTLSSettings{
					HttpsRedirect:  true,
					CredentialName: secretName,
				},
			})
	}

	gateway := &istiov1alpha3.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rio-gateway",
			Namespace: constants.IstioSystemNamespace,
		},
		Spec: networkingv1alpha3.Gateway{
			Servers:  servers,
			Selector: labelSelectors,
		},
	}
	result = append(result, gateway)

	return result, status, nil
}

func (h handler) copySecret(namespace, name string) *v1.Secret {
	existingSecret, err := h.secrets.Get(namespace, name)
	if err != nil {
		return nil
	}
	secret := constructors.NewSecret(constants.IstioSystemNamespace, name, v1.Secret{
		Data: existingSecret.Data,
	})
	return secret
}
