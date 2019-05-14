package issuer

import (
	"context"
	"fmt"
	"strings"

	"github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	certmanagerapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	"github.com/rancher/rio/modules/system/features/letsencrypt/pkg/issuers"
	v1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	projectv1controller "github.com/rancher/rio/pkg/generated/controllers/project.rio.cattle.io/v1"
	v12 "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/kv"
	"github.com/rancher/wrangler/pkg/objectset"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	rdnsTokenName   = "rdns-token"
	featureName     = "letsencrypt"
	rioWildcardType = "RIO_WILDCARD_CERT_TYPE"
	rdnsSuffix      = "on-rio.io"
)

func Register(ctx context.Context, rContext *types.Context) error {
	fh := &certsHandler{
		namespace: rContext.Namespace,
		apply: rContext.Apply.WithSetID("letsencrypt-issuer").
			WithStrictCaching().
			WithCacheTypes(rContext.CertManager.Certmanager().V1alpha1().ClusterIssuer(),
				rContext.CertManager.Certmanager().V1alpha1().Certificate()),
		clusterDomain: rContext.Global.Project().V1().ClusterDomain(),
		publicdomain:  rContext.Rio.Rio().V1().PublicDomain(),
		featureCache:  rContext.Global.Project().V1().Feature().Cache(),
	}

	rContext.Global.Project().V1().ClusterDomain().OnChange(ctx, "letsencrypt-clusterdomain-certs", fh.onChangeClusterDomain)
	rContext.CertManager.Certmanager().V1alpha1().Certificate().OnChange(ctx, "letsencrypt-certificate", fh.onChangeCert)
	return nil
}

type certsHandler struct {
	namespace     string
	apply         apply.Apply
	clusterDomain projectv1controller.ClusterDomainController
	publicdomain  v12.PublicDomainController
	featureCache  projectv1controller.FeatureCache
}

func (f *certsHandler) onChangeClusterDomain(key string, clusterDomain *v1.ClusterDomain) (*v1.ClusterDomain, error) {
	feature, err := f.featureCache.Get(f.namespace, featureName)
	if err != nil {
		return clusterDomain, err
	}

	domain := clusterDomain.Status.ClusterDomain
	if domain == "" {
		return clusterDomain, nil
	}

	os := objectset.NewObjectSet()
	for _, issuerName := range issuers.IssuerTypeToName {
		os.Add(constructIssuer(issuerName, domain))
	}

	if err := f.addWildcardCert(feature, domain, os); err != nil {
		return nil, err
	}
	return clusterDomain, f.apply.WithOwner(feature).Apply(os)
}

func (f *certsHandler) onChangeCert(key string, cert *v1alpha1.Certificate) (*v1alpha1.Certificate, error) {
	clusterDomain, err := f.clusterDomain.Cache().Get(f.namespace, constants.ClusterDomainName)
	if errors.IsNotFound(err) {
		return cert, nil
	} else if err != nil {
		return cert, err
	}

	if cert == nil || cert.Namespace != f.namespace {
		return cert, nil
	}

	if cert.Name == issuers.RioWildcardCerts {
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
					deepcopy.Status.HTTPSSupported = true

					_, err := f.publicdomain.Update(deepcopy)
					return cert, err
				}
			}
		}
	}

	return cert, nil
}

func (f *certsHandler) addWildcardCert(feature *v1.Feature, domain string, os *objectset.ObjectSet) error {
	if domain == "" || !strings.HasSuffix(domain, rdnsSuffix) {
		return nil
	}

	wildcardsType := feature.Spec.Answers[rioWildcardType]
	issuer := issuers.IssuerTypeToName[wildcardsType]
	if issuer == "" {
		return nil
	}

	os.Add(certificateDNS(f.namespace, issuers.RioWildcardCerts, domain, issuer))
	return nil
}

func constructIssuer(issuerName, domain string) *certmanagerapi.ClusterIssuer {
	issuer := constructors.NewClusterIssuer(issuerName, certmanagerapi.ClusterIssuer{})

	switch issuerName {
	case constants.StagingIssuerName, constants.ProductionIssuerName:
		acme := &certmanagerapi.ACMEIssuer{
			Email: emailFromDomain(domain),
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
		}
		if issuerName == constants.StagingIssuerName {
			acme.Server = constants.LetsEncryptStagingServerURL
		} else {
			acme.Server = constants.LetsEncryptProductionServerURL
		}
		issuer.Spec.ACME = acme
	case constants.SelfSignedIssuerName:
		issuer.Spec.SelfSigned = &certmanagerapi.SelfSignedIssuer{}
	}

	return issuer
}

func emailFromDomain(domain string) string {
	return fmt.Sprintf("user-%s@rancher.dev", strings.SplitN(domain, ".", 2)[0])
}

func certificateDNS(namespace, name, domain, issueName string) runtime.Object {
	wildcardDomain := "*." + domain
	return constructors.NewCertificate(namespace, name, certmanagerapi.Certificate{
		Spec: certmanagerapi.CertificateSpec{
			SecretName: issuers.RioWildcardCerts,
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
