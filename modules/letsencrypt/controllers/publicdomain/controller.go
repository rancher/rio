package publicdomain

import (
	"context"

	"github.com/rancher/wrangler/pkg/condition"
	"k8s.io/apimachinery/pkg/api/errors"

	v1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	adminv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/indexes"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/generic"
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
		certificateCache:       rContext.Admin.Admin().V1().Certificate().Cache(),
	}

	apply := rContext.Apply.
		WithCacheTypes(rContext.Admin.Admin().V1().Certificate())

	adminv1controller.RegisterPublicDomainGeneratingHandler(ctx,
		rContext.Admin.Admin().V1().PublicDomain(),
		apply,
		"LetsencryptCertificateDeployed",
		"letsencrypt-publicdomain",
		fh.Handle,
		&generic.GeneratingHandlerOptions{
			AllowClusterScoped: true,
		})

	rContext.Core.Core().V1().Secret().OnChange(ctx, "letsencrypt", fh.onSecretChange)

	return nil
}

type certsHandler struct {
	namespace              string
	secretsCache           corev1controller.SecretCache
	publicDomainCache      adminv1controller.PublicDomainCache
	publicDomainController adminv1controller.PublicDomainController
	certificateCache       adminv1controller.CertificateCache
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
	status.AssignedSecretName = cert.Spec.SecretRef.Name

	if status.AssignedSecretName == "" {
		status.HTTPSSupported = false
	} else {
		cert, err := f.certificateCache.Get(status.AssignedSecretName)
		if errors.IsNotFound(err) {
			status.HTTPSSupported = false
		} else if err != nil {
			return nil, status, err
		} else {
			status.HTTPSSupported = condition.Cond("CertificateDeployed").IsTrue(cert)
		}
	}

	return []runtime.Object{
		cert,
	}, status, nil
}

func certificateHTTP(namespace, domain string) *v1.Certificate {
	name := name2.SafeConcatName(domain, "tls")
	return &v1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1.CertificateSpec{
			SecretRef: corev1.SecretReference{
				Name:      name,
				Namespace: namespace,
			},
			DNSNames: []string{
				domain,
			},
		},
	}
}
