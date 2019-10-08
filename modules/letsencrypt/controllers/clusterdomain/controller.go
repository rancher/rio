package clusterdomain

import (
	"context"

	"github.com/rancher/rio/pkg/indexes"

	certmanagerapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	"github.com/rancher/rio/modules/letsencrypt/controllers/issuer"
	v1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	adminv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/certmanager.k8s.io/v1alpha1"
	"github.com/rancher/wrangler/pkg/condition"
	name2 "github.com/rancher/wrangler/pkg/name"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	fh := &certsHandler{
		namespace:               rContext.Namespace,
		certificateCache:        rContext.CertManager.Certmanager().V1alpha1().Certificate().Cache(),
		clusterDomainCache:      rContext.Admin.Admin().V1().ClusterDomain().Cache(),
		clusterDomainController: rContext.Admin.Admin().V1().ClusterDomain(),
	}

	apply := rContext.Apply.
		WithCacheTypes(rContext.CertManager.Certmanager().V1alpha1().Certificate())

	adminv1controller.RegisterClusterDomainGeneratingHandler(ctx,
		rContext.Admin.Admin().V1().ClusterDomain(),
		apply,
		"",
		"clusterdomain-letsencrypt",
		fh.Handle,
		nil)

	rContext.CertManager.Certmanager().V1alpha1().Certificate().OnChange(ctx, "letsencrypt", fh.onCertChange)

	return nil
}

type certsHandler struct {
	namespace               string
	certificateCache        v1alpha1.CertificateCache
	clusterDomainCache      adminv1controller.ClusterDomainCache
	clusterDomainController adminv1controller.ClusterDomainController
}

func (f *certsHandler) onCertChange(key string, obj *certmanagerapi.Certificate) (*certmanagerapi.Certificate, error) {
	if obj == nil {
		return nil, nil
	}
	domains, err := f.clusterDomainCache.GetByIndex(indexes.ClusterDomainByAssignedSecret, obj.Spec.SecretName)
	if err != nil {
		return obj, err
	}
	for _, domain := range domains {
		f.clusterDomainController.Enqueue(domain.Name)
	}
	return obj, nil
}

func (f *certsHandler) Handle(obj *v1.ClusterDomain, status v1.ClusterDomainStatus) ([]runtime.Object, v1.ClusterDomainStatus, error) {
	if obj.Spec.Provider != "rdns" {
		return nil, status, nil
	}

	if obj.Spec.SecretName != "" {
		status.AssignedSecretName = obj.Spec.SecretName
		return nil, status, nil
	}

	cert := wildcardDNS(f.namespace, obj.Name)
	status.AssignedSecretName = cert.Spec.SecretName

	if status.AssignedSecretName == "" {
		status.HTTPSSupported = false
	} else {
		cert, err := f.certificateCache.Get(f.namespace, status.AssignedSecretName)
		if errors.IsNotFound(err) {
			status.HTTPSSupported = false
		} else if err != nil {
			return nil, status, err
		} else {
			status.HTTPSSupported = condition.Cond("Ready").IsTrue(cert)
		}
	}

	return []runtime.Object{
		cert,
	}, status, nil
}

func wildcardDNS(namespace, name string) *certmanagerapi.Certificate {
	secretName := name2.SafeConcatName(name + "-tls")
	wildcardDomain := "*." + name
	return &certmanagerapi.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      secretName,
		},
		Spec: certmanagerapi.CertificateSpec{
			SecretName: secretName,
			IssuerRef: certmanagerapi.ObjectReference{
				Kind: "Issuer",
				Name: issuer.RioIssuer,
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
}
