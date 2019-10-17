package publicdomain

import (
	"context"

	"github.com/rancher/rio/pkg/indexes"

	certmanagerapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	"github.com/rancher/rio/modules/letsencrypt/controllers/issuer"
	v1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	adminv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	name2 "github.com/rancher/wrangler/pkg/name"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	fh := &certsHandler{
		namespace:              rContext.Namespace,
		secretsCache:           rContext.Core.Core().V1().Secret().Cache(),
		publicDomainCache:      rContext.Admin.Admin().V1().PublicDomain().Cache(),
		publicDomainController: rContext.Admin.Admin().V1().PublicDomain(),
	}

	apply := rContext.Apply.
		WithCacheTypes(rContext.CertManager.Certmanager().V1alpha1().Certificate())

	adminv1controller.RegisterPublicDomainGeneratingHandler(ctx,
		rContext.Admin.Admin().V1().PublicDomain(),
		apply,
		"",
		"letsencrypt-publicdomain",
		fh.Handle,
		nil)

	rContext.Core.Core().V1().Secret().OnChange(ctx, "letsencrypt", fh.onSecretChange)

	return nil
}

type certsHandler struct {
	namespace              string
	secretsCache           corev1controller.SecretCache
	publicDomainCache      adminv1controller.PublicDomainCache
	publicDomainController adminv1controller.PublicDomainController
}

func (f *certsHandler) onSecretChange(key string, obj *corev1.Secret) (*corev1.Secret, error) {
	domains, err := f.publicDomainCache.GetByIndex(indexes.PublicDomainByAssignedSecret, key)
	if err != nil {
		return obj, err
	}
	for _, domain := range domains {
		f.publicDomainController.Enqueue(domain.Name)
	}
	return obj, nil
}

func (f *certsHandler) Handle(obj *v1.PublicDomain, status v1.PublicDomainStatus) ([]runtime.Object, v1.PublicDomainStatus, error) {
	if obj.Namespace != f.namespace {
		return nil, status, nil
	}

	if obj.Spec.SecretName != "" {
		status.AssignedSecretName = obj.Spec.SecretName
		return nil, status, nil
	}

	cert := certificateHTTP(obj.Namespace, obj.Name)
	status.AssignedSecretName = cert.Spec.SecretName

	if status.AssignedSecretName == "" {
		status.HTTPSSupported = false
	} else {
		_, err := f.secretsCache.Get(obj.Namespace, status.AssignedSecretName)
		status.HTTPSSupported = err == nil
	}

	return []runtime.Object{
		cert,
	}, status, nil
}

func certificateHTTP(namespace, domain string) *certmanagerapi.Certificate {
	name := name2.SafeConcatName(domain, "-tls")
	return &certmanagerapi.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: certmanagerapi.CertificateSpec{
			SecretName: name,
			IssuerRef: certmanagerapi.ObjectReference{
				Kind: "Issuer",
				Name: issuer.RioIssuer,
			},
			DNSNames: []string{
				domain,
			},
			ACME: &certmanagerapi.ACMECertificateConfig{
				Config: []certmanagerapi.DomainSolverConfig{
					{
						Domains: []string{
							domain,
						},
						SolverConfig: certmanagerapi.SolverConfig{
							HTTP01: &certmanagerapi.HTTP01SolverConfig{},
						},
					},
				},
			},
		},
	}
}
