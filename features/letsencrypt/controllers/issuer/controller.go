package issuer

import (
	"context"

	certmanagerapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/features/letsencrypt/pkg/issuers"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/certmanager.k8s.io/v1alpha1"
	"github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	rdnsTokenName    = "rdns-token"
	TLSSecretName    = "rio-certs"
	rioWildcardCerts = "rio-wildcard"
)

func Register(ctx context.Context, rContext *types.Context) error {
	fh := &featureHandler{
		processor: objectset.NewProcessor("letsencrypt-issuer").
			Client(rContext.CertManager.ClusterIssuer,
				rContext.CertManager.Certificate),
	}

	rContext.Global.Feature.OnChange(ctx, "letsencrypt-issuer-controller", fh.onChange)
	return nil
}

type featureHandler struct {
	processor objectset.Processor
}

func (f *featureHandler) onChange(feature *v1.Feature) (runtime.Object, error) {
	if feature.Namespace != settings.RioSystemNamespace || feature.Name != "letsencrypt" {
		return feature, nil
	}

	os := objectset.NewObjectSet()
	for _, issuerName := range issuers.IssuerTypeToName {
		os.Add(constructIssuer(issuerName))
	}

	f.addWildcardCert(feature, os)

	return feature, f.processor.NewDesiredSet(feature, os).Apply()
}

func (f *featureHandler) addWildcardCert(feature *v1.Feature, os *objectset.ObjectSet) {
	if settings.ClusterDomain.Get() == "" {
		return
	}

	wildcardsType := feature.Spec.Answers[settings.RioWildcardType]
	issuer := issuers.IssuerTypeToName[wildcardsType]
	if issuer == "" {
		return
	}

	os.Add(certificateDNS(rioWildcardCerts, settings.ClusterDomain.Get(), issuer))
}

func constructIssuer(issuerName string) *certmanagerapi.ClusterIssuer {
	issuer := v1alpha1.NewClusterIssuer("", issuerName, certmanagerapi.ClusterIssuer{})

	switch issuerName {
	case settings.StagingIssuerName, settings.ProductionIssuerName:
		acme := &certmanagerapi.ACMEIssuer{
			Email: settings.LetsEncryptAccountEmail.Get(),
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
							APIEndpoint: settings.RDNSURL.Get(),
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
			acme.Server = settings.LetsEncryptStagingServerUrl.Get()
		} else {
			acme.Server = settings.LetsEncryptProductionServerUrl.Get()
		}
		issuer.Spec.ACME = acme
	case settings.SelfSignedIssuerName:
		issuer.Spec.SelfSigned = &certmanagerapi.SelfSignedIssuer{}
	}

	return issuer
}

func certificateDNS(name, domain, issueName string) runtime.Object {
	wildcardDomain := "*." + domain
	return v1alpha1.NewCertificate(settings.RioSystemNamespace, name, certmanagerapi.Certificate{
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
