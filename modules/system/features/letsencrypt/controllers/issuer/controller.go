package issuer

import (
	"context"

	"github.com/rancher/wrangler/pkg/kv"

	"github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	certmanagerapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	"github.com/rancher/rio/exclude/pkg/settings"
	"github.com/rancher/rio/modules/system/features/letsencrypt/pkg/issuers"
	v1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	projectv1controller "github.com/rancher/rio/pkg/generated/controllers/project.rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/objectset"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	rdnsTokenName    = "rdns-token"
	TLSSecretName    = "rio-certs"
	rioWildcardCerts = "rio-wildcard"
	featureName      = "letsencrypt"
	rioWildcardType  = "RIO_WILDCARD_CERT_TYPE"
)

func Register(ctx context.Context, rContext *types.Context) error {
	fh := &featureHandler{
		namespace: rContext.Namespace,
		apply: rContext.Apply.WithSetID("letsencrypt-issuer").
			WithStrictCaching().
			WithCacheTypes(rContext.CertManager.Certmanager().V1alpha1().ClusterIssuer(),
				rContext.CertManager.Certmanager().V1alpha1().Certificate()),
		clusterDomain: rContext.Global.Project().V1().ClusterDomain(),
		publicdomain:  rContext.Global.Project().V1().PublicDomain(),
	}

	rContext.Global.Project().V1().Feature().OnChange(ctx, "letsencrypt-issuer-controller", fh.onChange)
	rContext.CertManager.Certmanager().V1alpha1().Certificate().OnChange(ctx, "letsencrypt-certificate", fh.onChangeCert)
	return nil
}

type featureHandler struct {
	namespace     string
	apply         apply.Apply
	clusterDomain projectv1controller.ClusterDomainController
	publicdomain  projectv1controller.PublicDomainController
}

func (f *featureHandler) onChange(key string, feature *v1.Feature) (*v1.Feature, error) {
	if feature == nil || featureName != feature.Name || f.namespace != feature.Namespace {
		return nil, nil
	}

	os := objectset.NewObjectSet()
	for _, issuerName := range issuers.IssuerTypeToName {
		os.Add(constructIssuer(issuerName))
	}

	if err := f.addWildcardCert(feature, os); err != nil {
		return nil, err
	}

	return feature, f.apply.WithOwner(feature).Apply(os)
}

func (f *featureHandler) onChangeCert(key string, cert *v1alpha1.Certificate) (*v1alpha1.Certificate, error) {
	clusterDomain, err := f.clusterDomain.Cache().Get(f.namespace, settings.ClusterDomainName)
	if errors.IsNotFound(err) {
		return cert, nil
	} else if err != nil {
		return cert, err
	}

	if cert == nil || cert.Namespace != f.namespace {
		return cert, nil
	}

	if cert.Name == rioWildcardCerts {
		for _, con := range cert.Status.Conditions {
			if con.Type == v1alpha1.CertificateConditionReady && con.Status == certmanagerapi.ConditionTrue {
				deepcopy := clusterDomain.DeepCopy()
				deepcopy.Status.HTTPSSupported = true
				if _, err := f.clusterDomain.Update(deepcopy); err != nil {
					return cert, err
				}
				break
			}
		}
	}

	ns, name := kv.Split(cert.Name, "/")
	if ns != "" && name != "" {
		for _, con := range cert.Status.Conditions {
			if con.Type == v1alpha1.CertificateConditionReady && con.Status == certmanagerapi.ConditionTrue {
				// update public domain
				publicDomain, err := f.publicdomain.Cache().Get(ns, name)
				if err == nil {
					deepcopy := publicDomain.DeepCopy()
					deepcopy.Status.HttpsSupported = true
					_, err := f.publicdomain.Update(deepcopy)
					return cert, err
				}

				// todo: update gateway with secret ref
			}
		}
	}

	return cert, nil
}

func (f *featureHandler) addWildcardCert(feature *v1.Feature, os *objectset.ObjectSet) error {
	domain, err := f.getClusterDomain()
	if err != nil {
		return err
	}

	if domain == "" {
		return nil
	}

	wildcardsType := feature.Spec.Answers[rioWildcardType]
	issuer := issuers.IssuerTypeToName[wildcardsType]
	if issuer == "" {
		return nil
	}

	os.Add(certificateDNS(f.namespace, rioWildcardCerts, domain, issuer))
	return nil
}

func (f *featureHandler) getClusterDomain() (string, error) {
	clusterDomain, err := f.clusterDomain.Cache().Get(f.namespace, settings.ClusterDomainName)
	if errors.IsNotFound(err) {
		return "", nil
	} else if err != nil {
		return "", err
	}

	return clusterDomain.Status.ClusterDomain, nil
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
