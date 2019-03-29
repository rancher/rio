package issuer

import (
	"context"

	certmanagerapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	"github.com/rancher/rio/features/letsencrypt/pkg/issuers"
	v1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/objectset"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	rdnsTokenName    = "rdns-token"
	TLSSecretName    = "rio-certs"
	rioWildcardCerts = "rio-wildcard"
)

func Register(ctx context.Context, rContext *types.Context) error {
	fh := &featureHandler{
		namespace: rContext.Namespace,
		apply: rContext.Apply.WithSetID("letsencrypt-issuer").
			WithStrictCaching().
			WithCacheTypes(rContext.CertManager.Certmanager().V1alpha1().ClusterIssuer(),
				rContext.CertManager.Certmanager().V1alpha1().Certificate()),
	}

	rContext.Global.Project().V1().Feature().OnChange(ctx, "letsencrypt-issuer-controller", fh.onChange)
	return nil
}

type featureHandler struct {
	namespace string
	apply     apply.Apply
}

func (f *featureHandler) onChange(key string, feature *v1.Feature) (*v1.Feature, error) {
	if feature == nil {
		return nil, nil
	}

	if feature.Namespace != f.namespace || feature.Name != "letsencrypt" {
		return feature, nil
	}

	os := objectset.NewObjectSet()
	for _, issuerName := range issuers.IssuerTypeToName {
		os.Add(constructIssuer(issuerName))
	}

	f.addWildcardCert(feature, os)

	return feature, f.apply.WithOwner(feature).Apply(os)
}

func (f *featureHandler) addWildcardCert(feature *v1.Feature, os *objectset.ObjectSet) {
	if settings.ClusterDomain == "" {
		return
	}

	wildcardsType := feature.Spec.Answers[settings.RioWildcardType]
	issuer := issuers.IssuerTypeToName[wildcardsType]
	if issuer == "" {
		return
	}

	os.Add(certificateDNS(f.namespace, rioWildcardCerts, settings.ClusterDomain, issuer))
}

func constructIssuer(issuerName string) *certmanagerapi.ClusterIssuer {
	issuer := constructors.NewClusterIssuer(issuerName, certmanagerapi.ClusterIssuer{})

	switch issuerName {
	case settings.StagingIssuerName, settings.ProductionIssuerName:
		acme := &certmanagerapi.ACMEIssuer{
			Email: settings.LetsEncryptAccountEmail,
			PrivateKey: certmanagerapi.SecretKeySelector{
				LocalObjectReference: certmanagerapi.LocalObjectReference{
					Name: "letsencrypt-account",
				},
			},
			HTTP01: &certmanagerapi.ACMEIssuerHTTP01Config{},
			DNS01: &certmanagerapi.ACMEIssuerDNS01Config{
				Providers: []certmanagerapi.ACMEIssuerDNS01Provider{
					{
						Name: "rdns",
						RDNS: &certmanagerapi.ACMEIssuerDNS01ProviderRDNS{
							APIEndpoint: settings.RDNSURL,
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
		}
		if issuerName == settings.StagingIssuerName {
			acme.Server = settings.LetsEncryptStagingServerUrl
		} else {
			acme.Server = settings.LetsEncryptProductionServerUrl
		}
		issuer.Spec.ACME = acme
	case settings.SelfSignedIssuerName:
		issuer.Spec.SelfSigned = &certmanagerapi.SelfSignedIssuer{}
	}

	return issuer
}

func certificateDNS(namespace, name, domain, issueName string) runtime.Object {
	wildcardDomain := "*." + domain
	return constructors.NewCertificate(namespace, name, certmanagerapi.Certificate{
		Spec: certmanagerapi.CertificateSpec{
			SecretName: TLSSecretName,
			IssuerRef: certmanagerapi.ObjectReference{
				Kind: "ClusterIssuer",
				Name: issueName,
			},
			DNSNames: []string{
				wildcardDomain,
			},
			ACME: &certmanagerapi.ACMECertificateConfig{
				Config: []certmanagerapi.DomainSolverConfig{
					{
						Domains: []string{
							wildcardDomain,
						},
						SolverConfig: certmanagerapi.SolverConfig{
							DNS01: &certmanagerapi.DNS01SolverConfig{
								Provider: "rdns",
							},
						},
					},
				},
			},
		},
	})
}
