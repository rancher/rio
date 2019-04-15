package publicdomain

import (
	"context"
	"fmt"

	certmanagerapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	"github.com/rancher/rio/modules/system/features/letsencrypt/pkg/issuers"
	projectv1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	v12 "github.com/rancher/rio/pkg/generated/controllers/project.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/objectset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContexts *types.Context) error {
	p := &publicDomainHandler{
		namespace: rContexts.Namespace,
		apply: rContexts.Apply.WithSetID("letsencrypt-publicdomain").
			WithStrictCaching().
			WithCacheTypes(rContexts.CertManager.Certmanager().V1alpha1().Certificate()),
		featureClientCache: rContexts.Global.Project().V1().Feature().Cache(),
		publicDomains:      rContexts.Rio.Rio().V1().PublicDomain(),
		publicDomainCache:  rContexts.Rio.Rio().V1().PublicDomain().Cache(),
	}

	rContexts.Rio.Rio().V1().PublicDomain().OnChange(ctx, "letsencrypt-handler", p.onChange)
	rContexts.Rio.Rio().V1().PublicDomain().OnRemove(ctx, "letsencrypt-handler", p.onRemove)
	rContexts.Global.Project().V1().Feature().OnChange(ctx, "letsencrypt-handler", p.featureChanged)

	return nil
}

type publicDomainHandler struct {
	namespace          string
	apply              apply.Apply
	publicDomains      riov1controller.PublicDomainController
	publicDomainCache  riov1controller.PublicDomainCache
	featureClientCache v12.FeatureCache
}

func (p *publicDomainHandler) featureChanged(key string, feature *projectv1.Feature) (*projectv1.Feature, error) {
	if feature == nil {
		return feature, nil
	}

	if feature.Namespace != p.namespace || feature.Name != "letsencrypt" {
		return feature, nil
	}

	pds, err := p.publicDomainCache.List(p.namespace, labels.Everything())
	if err != nil {
		return feature, err
	}

	for _, pd := range pds {
		p.publicDomains.Enqueue(pd.Namespace, pd.Name)
	}

	return feature, nil
}

func (p *publicDomainHandler) onChange(key string, domain *riov1.PublicDomain) (*riov1.PublicDomain, error) {
	if domain == nil {
		return nil, nil
	}

	domain.Status.Endpoint = fmt.Sprintf("https://%s", domain.Spec.DomainName)

	// if public domain have secretRef, do not configure letsencrypt secrets
	if domain.Spec.SecretRef.Name != "" {
		return domain, nil
	}

	feature, err := p.featureClientCache.Get(p.namespace, "letsencrypt")
	if err != nil {
		return domain, err
	}

	publicdomainType := feature.Spec.Answers[settings.PublicDomainType]
	issuerName := issuers.IssuerTypeToName[publicdomainType]

	os := objectset.NewObjectSet()

	if issuerName != "" {
		os.Add(certificateHttp(p.namespace, domain, issuerName))
	}

	return domain, p.apply.WithOwner(domain).Apply(os)
}

func (p *publicDomainHandler) onRemove(key string, domain *riov1.PublicDomain) (*riov1.PublicDomain, error) {
	if domain == nil {
		return nil, nil
	}

	if domain.Namespace != p.namespace {
		return domain, nil
	}

	return domain, p.apply.WithOwner(domain).Apply(nil)
}

func certificateHttp(namespace string, domain *riov1.PublicDomain, issuerName string) runtime.Object {
	name := fmt.Sprintf("%s-%s", domain.Namespace, domain.Name)
	cert := constructors.NewCertificate(namespace, name,
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
