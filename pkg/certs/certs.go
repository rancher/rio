package certs

import (
	"fmt"

	certmanagerapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	"github.com/rancher/rio/pkg/apply"
	"github.com/rancher/rio/pkg/deploy/istio/output"
	"github.com/rancher/rio/pkg/settings"
	projectv1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	TlsSecretName    = "rio-certs"
	RioWildcardCerts = "rio-wildcard"
	RdnsTokenName    = "rdns-token"
)

func ApplyWildcardCertificates() error {
	if err := apply.Apply([]runtime.Object{AcmeIssuer()}, nil, "", "acme-cluster-issuer"); err != nil {
		return err
	}
	domain := settings.ClusterDomain.Get()
	if domain == "" {
		return nil
	}
	return apply.Apply([]runtime.Object{CertificateDNS(RioWildcardCerts, domain)}, nil, settings.RioSystemNamespace, "certificate-wildcard-dns")
}

func CertificateHttp(domain *projectv1.PublicDomain) runtime.Object {
	cert := &output.Certificate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "certmanager.k8s.io/v1alpha1",
			Kind:       "Certificate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-tls-certs", domain.Name),
			Namespace: settings.RioSystemNamespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: domain.APIVersion,
					Kind:       domain.Kind,
					Name:       domain.Name,
					UID:        domain.UID,
				},
			},
		},
	}

	certSpec := certmanagerapi.CertificateSpec{
		SecretName: cert.Name,
		IssuerRef: certmanagerapi.ObjectReference{
			Kind: "ClusterIssuer",
			Name: settings.CerManagerIssuerName,
		},
		DNSNames: []string{
			domain.Spec.DomainName,
		},
		ACME: &certmanagerapi.ACMECertificateConfig{
			Config: []certmanagerapi.DomainSolverConfig{
				{
					Domains: []string{
						domain.Spec.DomainName,
					},
					SolverConfig: certmanagerapi.SolverConfig{
						HTTP01: &certmanagerapi.HTTP01SolverConfig{},
					},
				},
			},
		},
	}
	cert.Spec = certSpec
	return cert
}

func CertificateDNS(name, domain string) runtime.Object {
	wildcardDomain := "*." + domain
	cert := &output.Certificate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "certmanager.k8s.io/v1alpha1",
			Kind:       "Certificate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s", name),
			Namespace: settings.RioSystemNamespace,
		},
	}

	certSpec := certmanagerapi.CertificateSpec{
		SecretName: TlsSecretName,
		IssuerRef: certmanagerapi.ObjectReference{
			Kind: "ClusterIssuer",
			Name: settings.CerManagerIssuerName,
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
	}
	cert.Spec = certSpec
	return cert
}

func AcmeIssuer() runtime.Object {
	clusterIssuer := &output.ClusterIssuer{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterIssuer",
			APIVersion: "certmanager.k8s.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: settings.CerManagerIssuerName,
		},
	}
	issuerSpec := &certmanagerapi.IssuerSpec{
		IssuerConfig: certmanagerapi.IssuerConfig{
			ACME: &certmanagerapi.ACMEIssuer{
				Email:  settings.LetsEncryptAccountEmail.Get(),
				Server: settings.LetsEncryptServerUrl.Get(),
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
										Name: RdnsTokenName,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	clusterIssuer.Spec = issuerSpec
	return clusterIssuer
}
