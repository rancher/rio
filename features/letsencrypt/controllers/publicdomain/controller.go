package publicdomain

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/labels"

	certmanagerapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/features/letsencrypt/pkg/issuers"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/certmanager.k8s.io/v1alpha1"
	"github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContexts *types.Context) error {
	p := &publicDomainHandler{
		processor: objectset.NewProcessor("letsencrypt-publicdomain").
			Client(rContexts.CertManager.Certificate),
		featureClientCache: rContexts.Global.Feature.Cache(),
		publicDomains:      rContexts.Global.PublicDomain,
		publicDomainCache:  rContexts.Global.PublicDomain.Cache(),
	}

	rContexts.Global.PublicDomain.OnChange(ctx, "letsencrypt-handler", p.onChange)
	rContexts.Global.PublicDomain.OnRemove(ctx, "letsencrypt-handler", p.onRemove)
	rContexts.Global.Feature.Interface().AddHandler(ctx, "letsencrypt-handler", p.featureChanged)

	return nil
}

type publicDomainHandler struct {
	processor          objectset.Processor
	publicDomains      v1.PublicDomainClient
	publicDomainCache  v1.PublicDomainClientCache
	featureClientCache v1.FeatureClientCache
}

func (p *publicDomainHandler) featureChanged(key string, feature *v1.Feature) (runtime.Object, error) {
	if feature == nil {
		return feature, nil
	}

	if feature.Namespace != settings.RioSystemNamespace || feature.Name != "letsencrypt" {
		return feature, nil
	}

	pds, err := p.publicDomainCache.List(settings.RioSystemNamespace, labels.Everything())
	if err != nil {
		return feature, err
	}

	for _, pd := range pds {
		p.publicDomains.Enqueue(pd.Namespace, pd.Name)
	}

	return feature, nil
}

func (p *publicDomainHandler) onChange(domain *v1.PublicDomain) (runtime.Object, error) {
	if domain.Namespace != settings.RioSystemNamespace {
		return domain, nil
	}

	feature, err := p.featureClientCache.Get(settings.RioSystemNamespace, "letsencrypt")
	if err != nil {
		return domain, err
	}

	publicdomainType := feature.Spec.Answers[settings.PublicDomainType]
	issuerName := issuers.IssuerTypeToName[publicdomainType]

	os := objectset.NewObjectSet()

	if issuerName == "" {
		os.Add(certificateHttp(domain, issuerName))
	}

	return domain, p.processor.NewDesiredSet(domain, os).Apply()
}

func (p *publicDomainHandler) onRemove(domain *v1.PublicDomain) (runtime.Object, error) {
	if domain.Namespace != settings.RioSystemNamespace {
		return domain, nil
	}

	return domain, p.processor.Remove(domain)
}

func certificateHttp(domain *v1.PublicDomain, issuerName string) runtime.Object {
	name := fmt.Sprintf("%s-tls-certs", domain.Name)
	cert := v1alpha1.NewCertificate(settings.RioSystemNamespace, name,
		certmanagerapi.Certificate{
			ObjectMeta: metav1.ObjectMeta{
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
				SecretName: name,
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
		})
	return cert
}
