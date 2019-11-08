package publicdomain

import (
	"context"

	certmanagerv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"github.com/rancher/rio/modules/letsencrypt/controllers/issuer"
	v1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	adminv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/indexes"
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
		WithCacheTypes(rContext.CertManager.Certmanager().V1alpha2().Certificate())

	adminv1controller.RegisterPublicDomainGeneratingHandler(ctx,
		rContext.Admin.Admin().V1().PublicDomain(),
		apply,
		"LetsencryptCertificateDeployed",
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
	if obj.Spec.SecretName != "" {
		status.AssignedSecretName = obj.Spec.SecretName
		return nil, status, nil
	}

	cert := certificateHTTP(f.namespace, obj.Name)
	status.AssignedSecretName = cert.Spec.SecretName

	if status.AssignedSecretName == "" {
		status.HTTPSSupported = false
	} else {
		_, err := f.secretsCache.Get(f.namespace, status.AssignedSecretName)
		status.HTTPSSupported = err == nil
	}

	return []runtime.Object{
		cert,
	}, status, nil
}

func certificateHTTP(namespace, domain string) *certmanagerv1alpha2.Certificate {
	name := name2.SafeConcatName(domain, "tls")
	return &certmanagerv1alpha2.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Annotations: map[string]string{
				"cert-manager.io/issue-temporary-certificate": "true",
			},
		},
		Spec: certmanagerv1alpha2.CertificateSpec{
			SecretName: name,
			IssuerRef: cmmeta.ObjectReference{
				Kind: "Issuer",
				Name: issuer.RioHTTPIssuer,
			},
			DNSNames: []string{
				domain,
			},
		},
	}
}
