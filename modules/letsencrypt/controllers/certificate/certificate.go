package certificate

import (
	"context"
	"strings"

	"github.com/go-acme/lego/challenge/http01"

	"github.com/go-acme/lego/v3/certcrypto"
	"github.com/go-acme/lego/v3/certificate"
	"github.com/go-acme/lego/v3/lego"
	"github.com/rancher/rio/modules/letsencrypt/pkg"
	v1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	adminv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/apply"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	dnsType  = "dns-01"
	httpType = "http-01"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		namespace: rContext.Namespace,
		apply:     rContext.Apply.WithCacheTypes(rContext.Core.Core().V1().Secret()),
		secrets:   rContext.Core.Core().V1().Secret().Cache(),
	}

	adminv1controller.RegisterCertificateGeneratingHandler(ctx,
		rContext.Admin.Admin().V1().Certificate(),
		rContext.Apply.WithCacheTypes(rContext.Core.Core().V1().Secret()),
		"CertificateDeployed",
		"certificate-provisioned",
		h.generate,
		nil,
	)

	return nil
}

type handler struct {
	namespace string
	apply     apply.Apply
	secrets   corev1controller.SecretCache
}

func (h handler) generate(obj *v1.Certificate, status v1.CertificateStatus) ([]runtime.Object, v1.CertificateStatus, error) {
	if obj == nil || obj.DeletionTimestamp != nil {
		return nil, status, nil
	}

	providerType := ""
	for _, dnsName := range obj.Spec.DNSNames {
		if strings.HasPrefix(dnsName, "*") {
			providerType = dnsType
		} else {
			providerType = httpType
		}
		break
	}

	secret, err := h.secrets.Get(h.namespace, constants.LetsEncryptAccountSecretName)
	if err != nil {
		return nil, status, err
	}

	user, err := pkg.FromSecret(secret)
	if err != nil {
		return nil, status, err
	}

	config := lego.NewConfig(user)

	// This CA URL is configured for a local dev instance of Boulder running in Docker in a VM.
	config.CADirURL = user.URL
	config.Certificate.KeyType = certcrypto.RSA2048

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		return nil, status, err
	}

	switch providerType {
	case dnsType:
		rdnsSecret, err := h.secrets.Get(h.namespace, pkg.RdnsSecretName)
		if err != nil {
			return nil, status, err
		}

		rdnsProvider, err := pkg.NewDNSProviderCredential(constants.RDNSURL, string(rdnsSecret.Data["token"]))
		if err != nil {
			return nil, status, err
		}

		err = client.Challenge.SetDNS01Provider(rdnsProvider)
		if err != nil {
			return nil, status, err
		}

		request := certificate.ObtainRequest{
			Domains: obj.Spec.DNSNames,
			Bundle:  true,
		}
		certificates, err := client.Certificate.Obtain(request)
		if err != nil {
			return nil, status, err
		}

		secret := constructors.NewSecret(obj.Spec.SecretRef.Namespace, obj.Spec.SecretRef.Name, corev1.Secret{
			Type: corev1.SecretTypeTLS,
			Data: map[string][]byte{
				corev1.TLSPrivateKeyKey: certificates.PrivateKey,
				corev1.TLSCertKey:       certificates.Certificate,
			},
		})

		return []runtime.Object{secret}, status, nil
	case httpType:
		err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", "8080"))
		if err != nil {
			return nil, status, err
		}
		request := certificate.ObtainRequest{
			Domains: obj.Spec.DNSNames,
			Bundle:  true,
		}
		certificates, err := client.Certificate.Obtain(request)
		if err != nil {
			return nil, status, err
		}
		secret := constructors.NewSecret(obj.Spec.SecretRef.Namespace, obj.Spec.SecretRef.Name, corev1.Secret{
			Type: corev1.SecretTypeTLS,
			Data: map[string][]byte{
				corev1.TLSPrivateKeyKey: certificates.PrivateKey,
				corev1.TLSCertKey:       certificates.Certificate,
			},
		})

		return []runtime.Object{secret}, status, nil
	}

	return nil, status, err
}
