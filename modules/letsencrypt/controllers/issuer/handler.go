package issuer

import (
	"context"
	"fmt"

	certmanagerapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	"github.com/rancher/rio/pkg/config"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/apply"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	handlerName   = "letsencrypt-issuer"
	rdnsTokenName = "rdns-token"
	RioIssuer     = "rio-issuer"

	defaultEmail     = "cert@rancher.dev"
	defaultAccount   = "letsencrypt-account"
	defaultServerURL = "https://acme-v02.api.letsencrypt.org/directory"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := &handler{
		key:       fmt.Sprintf("%s/%s", rContext.Namespace, config.ConfigName),
		namespace: rContext.Namespace,
		apply: rContext.Apply.
			WithSetID(handlerName).
			WithSetOwnerReference(true, true).
			WithCacheTypes(
				rContext.CertManager.Certmanager().V1alpha1().Issuer()),
	}

	rContext.Core.Core().V1().ConfigMap().OnChange(ctx, handlerName, h.sync)
	return nil
}

type handler struct {
	key       string
	namespace string
	apply     apply.Apply
}

func (h *handler) sync(key string, cm *v1.ConfigMap) (*v1.ConfigMap, error) {
	if cm == nil || key != h.key {
		return nil, nil
	}

	config, err := config.FromConfigMap(cm)
	if err != nil {
		return cm, err
	}

	return cm, h.apply.
		WithOwner(cm).
		ApplyObjects(constructIssuer(h.namespace, config))
}

func withDefault(val, def string) string {
	if val == "" {
		return def
	}
	return val
}

func constructIssuer(namespace string, config config.Config) *certmanagerapi.Issuer {
	account := withDefault(config.LetsEncrypt.Account, defaultAccount)
	email := withDefault(config.LetsEncrypt.Email, defaultEmail)
	url := withDefault(config.LetsEncrypt.ServerURL, defaultServerURL)

	return &certmanagerapi.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      RioIssuer,
		},
		Spec: certmanagerapi.IssuerSpec{
			IssuerConfig: certmanagerapi.IssuerConfig{
				ACME: &certmanagerapi.ACMEIssuer{
					Server: url,
					Email:  email,
					PrivateKey: certmanagerapi.SecretKeySelector{
						LocalObjectReference: certmanagerapi.LocalObjectReference{
							Name: account,
						},
					},
					HTTP01: &certmanagerapi.ACMEIssuerHTTP01Config{},
					DNS01: &certmanagerapi.ACMEIssuerDNS01Config{
						Providers: []certmanagerapi.ACMEIssuerDNS01Provider{
							{
								Name: "rdns",
								RDNS: &certmanagerapi.ACMEIssuerDNS01ProviderRDNS{
									APIEndpoint: constants.RDNSURL,
									ClientToken: certmanagerapi.SecretKeySelector{
										Key: "token",
										LocalObjectReference: certmanagerapi.LocalObjectReference{
											Name: rdnsTokenName,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
