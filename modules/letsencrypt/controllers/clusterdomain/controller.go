package clusterdomain

import (
	"context"

	certmanagerv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"github.com/rancher/rio/modules/letsencrypt/controllers/issuer"
	v1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	adminv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/indexes"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/cert-manager.io/v1alpha2"
	"github.com/rancher/wrangler/pkg/condition"
	name2 "github.com/rancher/wrangler/pkg/name"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	fh := &certsHandler{
		namespace:               rContext.Namespace,
		certificateCache:        rContext.CertManager.Certmanager().V1alpha2().Certificate().Cache(),
		clusterDomainCache:      rContext.Admin.Admin().V1().ClusterDomain().Cache(),
		clusterDomainController: rContext.Admin.Admin().V1().ClusterDomain(),
	}

	apply := rContext.Apply.
		WithCacheTypes(rContext.CertManager.Certmanager().V1alpha2().Certificate())

	adminv1controller.RegisterClusterDomainGeneratingHandler(ctx,
		rContext.Admin.Admin().V1().ClusterDomain(),
		apply,
		"LetsencryptCertificateDeployed",
		"clusterdomain-letsencrypt",
		fh.Handle,
		nil)

	rContext.CertManager.Certmanager().V1alpha2().Certificate().OnChange(ctx, "letsencrypt", fh.onCertChange)

	return nil
}

type certsHandler struct {
	namespace               string
	certificateCache        v1alpha2.CertificateCache
	clusterDomainCache      adminv1controller.ClusterDomainCache
	clusterDomainController adminv1controller.ClusterDomainController
}

func (f *certsHandler) onCertChange(key string, obj *certmanagerv1alpha2.Certificate) (*certmanagerv1alpha2.Certificate, error) {
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

func wildcardDNS(namespace, name string) *certmanagerv1alpha2.Certificate {
	secretName := name2.SafeConcatName(name, "tls")
	wildcardDomain := "*." + name
	return &certmanagerv1alpha2.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      secretName,
			Annotations: map[string]string{
				"cert-manager.io/issue-temporary-certificate": "true",
			},
		},
		Spec: certmanagerv1alpha2.CertificateSpec{
			SecretName: secretName,
			IssuerRef: cmmeta.ObjectReference{
				Kind: "Issuer",
				Name: issuer.RioDNSIssuer,
			},
			DNSNames: []string{
				wildcardDomain,
			},
		},
	}
}
