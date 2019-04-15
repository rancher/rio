package secretaggregation

import (
	"context"
	"fmt"

	"github.com/rancher/rio/pkg/constructors"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/rancher/rio/modules/system/features/letsencrypt/pkg/issuers"
	v1 "github.com/rancher/rio/pkg/generated/controllers/core/v1"
	v12 "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	secretsAggragation = "secret-aggregation"
)

func Register(ctx context.Context, rContext *types.Context) error {
	s := secretaggregation{
		namespace:          rContext.Namespace,
		secrets:            rContext.Core.Core().V1().Secret(),
		publicDomainsCache: rContext.Rio.Rio().V1().PublicDomain().Cache(),
	}

	rContext.Core.Core().V1().Secret().OnChange(ctx, secretsAggragation, s.sync)
	return nil
}

type secretaggregation struct {
	namespace          string
	secrets            v1.SecretController
	publicDomainsCache v12.PublicDomainCache
}

func (s secretaggregation) sync(key string, secret *corev1.Secret) (*corev1.Secret, error) {
	publicdomainSecrets := map[string]struct{}{}
	publicdomains, err := s.publicDomainsCache.List("", labels.NewSelector())
	if err != nil {
		return secret, err
	}
	for _, pd := range publicdomains {
		publicdomainSecrets[fmt.Sprintf("%s/%s", pd.Spec.SecretRef.Namespace, pd.Spec.SecretRef.Name)] = struct{}{}
	}

	if !shouldAggregate(publicdomainSecrets, secret) {
		return secret, nil
	}

	aggregatedSecret, err := s.ensureSecrets()
	if err != nil {
		return secret, err
	}

	if aggregatedSecret.Data == nil {
		aggregatedSecret.Data = make(map[string][]byte, 0)
	}
	pkey := fmt.Sprintf("%s-%s-tls.key", secret.Namespace, secret.Name)
	cert := fmt.Sprintf("%s-%s-tls.crt", secret.Namespace, secret.Name)
	aggregatedSecret.Data[pkey] = secret.Data["tls.key"]
	aggregatedSecret.Data[cert] = secret.Data["tls.crt"]
	_, err = s.secrets.Update(aggregatedSecret)
	return secret, err
}

func (s secretaggregation) ensureSecrets() (*corev1.Secret, error) {
	secret, err := s.secrets.Get(s.namespace, issuers.TLSSecretName, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	} else if errors.IsNotFound(err) {
		secret = constructors.NewSecret(s.namespace, issuers.TLSSecretName, corev1.Secret{})
		secret, err = s.secrets.Create(secret)
	}
	return secret, err
}

func shouldAggregate(publicdomainSecrets map[string]struct{}, secret *corev1.Secret) bool {
	if secret.Name == issuers.RioWildcardCerts {
		return true
	}

	_, ok := publicdomainSecrets[fmt.Sprintf("%s/%s", secret.Namespace, secret.Name)]
	return ok
}
