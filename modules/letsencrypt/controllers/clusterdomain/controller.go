package clusterdomain

import (
	"context"

	"github.com/rancher/wrangler/pkg/generic"

	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	adminv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/indexes"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/condition"
	name2 "github.com/rancher/wrangler/pkg/name"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	fh := &certsHandler{
		namespace:               rContext.Namespace,
		clusterDomainCache:      rContext.Admin.Admin().V1().ClusterDomain().Cache(),
		clusterDomainController: rContext.Admin.Admin().V1().ClusterDomain(),
		certificateCache:        rContext.Admin.Admin().V1().Certificate().Cache(),
	}

	adminv1controller.RegisterClusterDomainGeneratingHandler(ctx,
		rContext.Admin.Admin().V1().ClusterDomain(),
		rContext.Apply.WithCacheTypes(rContext.Admin.Admin().V1().Certificate()),
		"LetsencryptCertificateDeployed",
		"clusterdomain-letsencrypt",
		fh.Handle,
		&generic.GeneratingHandlerOptions{
			AllowClusterScoped: true,
		})

	rContext.Admin.Admin().V1().Certificate().OnChange(ctx, "letsencrypt", fh.onCertChange)

	return nil
}

type certsHandler struct {
	namespace               string
	certificateCache        adminv1controller.CertificateCache
	clusterDomainCache      adminv1controller.ClusterDomainCache
	clusterDomainController adminv1controller.ClusterDomainController
}

func (f *certsHandler) onCertChange(key string, obj *adminv1.Certificate) (*adminv1.Certificate, error) {
	if obj == nil {
		return nil, nil
	}
	domains, err := f.clusterDomainCache.GetByIndex(indexes.ClusterDomainByAssignedSecret, obj.Spec.SecretRef.Name)
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
		status.HTTPSSupported = true
		return nil, status, nil
	}

	cert := wildcardDNS(f.namespace, obj.Name)
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

func wildcardDNS(namespace, name string) *v1.Certificate {
	secretName := name2.SafeConcatName(name, "tls")
	wildcardDomain := "*." + name
	return &adminv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      secretName,
			Annotations: map[string]string{
				"cert-manager.io/issue-temporary-certificate": "true",
			},
		},
		Spec: v1.CertificateSpec{
			SecretRef: corev1.SecretReference{
				Name:      secretName,
				Namespace: namespace,
			},
			DNSNames: []string{
				wildcardDomain,
			},
		},
	}
}
