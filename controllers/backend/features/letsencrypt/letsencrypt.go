package letsencrypt

import (
	"fmt"
	"time"

	certmanagerapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	"github.com/pkg/errors"
	"github.com/rancher/rio/pkg/apply"
	"github.com/rancher/rio/pkg/data"
	"github.com/rancher/rio/pkg/settings"
	projectv1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	TlsSecretName    = "rio-certs"
	RioWildcardCerts = "rio-wildcard"
	RdnsTokenName    = "rdns-token"
)

var issuerTypeToName = map[string]string{
	settings.StagingType:    settings.StagingIssuerName,
	settings.ProductionType: settings.ProductionIssuerName,
	settings.SelfSignedType: settings.SelfSignedIssuerName,
}

type Wrapper struct {
	PublicdomainLister projectv1.PublicDomainClientCache
	Feature            *projectv1.Feature
	Stacks             v1.StackClient
}

func (w Wrapper) Reconcile() error {
	var stacks []runtime.Object
	if w.Feature.Spec.Enable {
		stacks = append(stacks, data.Stack("cert-manager", v1.StackSpec{
			DisableMesh:               true,
			Answers:                   w.Feature.Spec.Answers,
			EnableKubernetesResources: true,
		}))
	}
	empty := []string{}
	if len(stacks) == 0 {
		empty = []string{"stacks.rio.cattle.io"}
	}

	if err := apply.Apply(stacks, empty, settings.RioSystemNamespace, "rio-letsencrypt-stacks"); err != nil {
		return err
	}
	if !w.Feature.Spec.Enable {
		return nil
	}
	starttime := time.Now()
	interval := time.Second
	for time.Now().Sub(starttime) < time.Minute*5 {
		stack, err := w.Stacks.Get(settings.RioSystemNamespace, "cert-manager", metav1.GetOptions{})
		if err != nil {
			return err
		}
		if v1.StackConditionDeployed.IsTrue(stack) {
			break
		}
		time.Sleep(interval)
		interval *= 2
	}

	StagingType := settings.StagingType
	ProductionType := settings.ProductionType
	SelfSignedType := settings.SelfSignedType
	wildcardsType := w.Feature.Spec.Answers[settings.RioWildcardType]
	if wildcardsType != StagingType && wildcardsType != ProductionType && wildcardsType != SelfSignedType {
		return errors.Errorf("rio wildcards certificate type must be %s, %s or %s", StagingType, ProductionType, SelfSignedType)
	}
	publicdomainType := w.Feature.Spec.Answers[settings.PublicDomainType]
	if publicdomainType != StagingType && publicdomainType != ProductionType && publicdomainType != SelfSignedType {
		return errors.Errorf("rio publicdomain certificate type must be %s, %s or %s", StagingType, ProductionType, SelfSignedType)
	}
	// creating issuers
	issuers := make([]runtime.Object, 0)
	for _, t := range []string{StagingType, ProductionType, SelfSignedType} {
		issuers = append(issuers, ConstructIssuer(issuerTypeToName[t]))
	}
	if err := apply.Apply(issuers, nil, "", "letsencrypts-issuer"); err != nil {
		return err
	}

	// rio wildcards certs
	if settings.ClusterDomain.Get() != "" {
		if err := applyWildcardCertificates(settings.ClusterDomain.Get(), wildcardsType); err != nil {
			return err
		}
	}

	// public domain certs
	publicdomains, err := w.PublicdomainLister.List(settings.RioSystemNamespace, labels.Everything())
	if err != nil {
		return err
	}
	certs := make([]runtime.Object, 0)
	for _, publicdomain := range publicdomains {
		certs = append(certs, certificateHttp(publicdomain, issuerTypeToName[publicdomainType]))
	}
	return apply.Apply(certs, nil, settings.RioSystemNamespace, "certfificate-publicdomain-http")
}

func applyWildcardCertificates(domain string, issuerType string) error {
	return apply.Apply(CertificateDNS(RioWildcardCerts, domain, issuerTypeToName[issuerType]), nil, settings.RioSystemNamespace, "certificate-wildcard-dns")
}

func certificateHttp(domain *projectv1.PublicDomain, issuerName string) runtime.Object {
	cert := &certmanagerapi.Certificate{
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
		Spec: certmanagerapi.CertificateSpec{
			SecretName: fmt.Sprintf("%s-tls-certs", domain.Name),
			IssuerRef: certmanagerapi.ObjectReference{
				Kind: "ClusterIssuer",
				Name: issuerName,
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
		},
	}
	return cert
}

func CertificateDNS(name, domain, issueName string) []runtime.Object {
	wildcardDomain := "*." + domain
	cert := &certmanagerapi.Certificate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "certmanager.k8s.io/v1alpha1",
			Kind:       "Certificate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s", name),
			Namespace: settings.RioSystemNamespace,
		},
		Spec: certmanagerapi.CertificateSpec{
			SecretName: TlsSecretName,
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
	}
	return []runtime.Object{cert}
}

func ConstructIssuer(issuerName string) runtime.Object {
	issuer := &certmanagerapi.ClusterIssuer{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterIssuer",
			APIVersion: "certmanager.k8s.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: issuerName,
		},
		Spec: certmanagerapi.IssuerSpec{
			IssuerConfig: certmanagerapi.IssuerConfig{},
		},
	}
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
									Name: RdnsTokenName,
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
