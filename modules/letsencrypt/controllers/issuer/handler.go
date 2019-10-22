package issuer

import (
	"context"
	"fmt"

	cmacme "github.com/jetstack/cert-manager/pkg/apis/acme/v1alpha2"
	certmanagerv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
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
	RioDNSIssuer  = "rio-dns-issuer"
	RioHTTPIssuer = "rio-http-issuer"

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
				rContext.CertManager.Certmanager().V1alpha2().Issuer(),
				rContext.Core.Core().V1().Secret()),
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
		ApplyObjects(
			constructIssuer(h.namespace, "dns", config),
			constructIssuer(h.namespace, "http", config),
		)
}

func withDefault(val, def string) string {
	if val == "" {
		return def
	}
	return val
}

func constructIssuer(namespace, issuerType string, config config.Config) *certmanagerv1alpha2.Issuer {
	account := withDefault(config.LetsEncrypt.Account, defaultAccount)
	email := withDefault(config.LetsEncrypt.Email, defaultEmail)
	url := withDefault(config.LetsEncrypt.ServerURL, defaultServerURL)

	issuer := &certmanagerv1alpha2.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
		},
		Spec: certmanagerv1alpha2.IssuerSpec{
			IssuerConfig: certmanagerv1alpha2.IssuerConfig{
				ACME: &cmacme.ACMEIssuer{
					Server: url,
					Email:  email,
					PrivateKey: cmmeta.SecretKeySelector{
						LocalObjectReference: cmmeta.LocalObjectReference{
							Name: account,
						},
					},
				},
			},
		},
	}

	if issuerType == "dns" {
		issuer.Name = RioDNSIssuer
		issuer.Spec.ACME.Solvers = []cmacme.ACMEChallengeSolver{
			{
				DNS01: &cmacme.ACMEChallengeSolverDNS01{
					RDNS: &cmacme.ACMEIssuerDNS01ProviderRDNS{
						APIEndpoint: constants.RDNSURL,
						ClientToken: cmmeta.SecretKeySelector{
							Key: "token",
							LocalObjectReference: cmmeta.LocalObjectReference{
								Name: rdnsTokenName,
							},
						},
					},
				},
			}}
	} else {
		issuer.Name = RioHTTPIssuer
		issuer.Spec.ACME.Solvers = []cmacme.ACMEChallengeSolver{
			{
				HTTP01: &cmacme.ACMEChallengeSolverHTTP01{
					Ingress: &cmacme.ACMEChallengeSolverHTTP01Ingress{
						Class: &[]string{"gloo"}[0],
					},
				},
			},
		}
	}

	return issuer
}
