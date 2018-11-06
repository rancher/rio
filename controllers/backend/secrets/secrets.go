package secrets

import (
	"context"
	"fmt"
	"sync"

	"github.com/rancher/rio/pkg/certs"

	"github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/rancher/rio/pkg/settings"

	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"github.com/rancher/types/apis/core/v1"
	corev1 "k8s.io/api/core/v1"
)

func Register(ctx context.Context, rContext *types.Context) {
	c := secretController{
		services:      rContext.Rio.Services(""),
		serviceLister: rContext.Rio.Services("").Controller().Lister(),
		secretsLister: rContext.Core.Secrets("").Controller().Lister(),
		secrets:       rContext.Core.Secrets(""),
	}
	rContext.Core.Secrets("").AddHandler("tls-secrets", c.sync)
}

type secretController struct {
	services      v1beta1.ServiceInterface
	serviceLister v1beta1.ServiceLister
	secretsLister v1.SecretLister
	secrets       v1.SecretInterface
	lock          sync.Mutex
}

func (s *secretController) sync(key string, secret *corev1.Secret) error {
	if secret == nil || secret.DeletionTimestamp != nil {
		return nil
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	if secret.Annotations["certmanager.k8s.io/issuer-name"] == settings.CerManagerIssuerName && secret.Namespace == settings.RioSystemNamespace && secret.Data["tls.crt"] != nil {
		// copy the cert into istio namespace and rename key
		if exitsingSecret, err := s.secretsLister.Get(settings.IstioExternalLBNamespace, certs.TlsSecretName); err != nil {
			if !errors.IsNotFound(err) {
				return err
			} else {
				newSecret := &corev1.Secret{}
				newSecret.Name = secret.Name
				newSecret.Namespace = settings.IstioExternalLBNamespace
				newSecret.Data = make(map[string][]byte)
				for k, v := range secret.Data {
					newSecret.Data[fmt.Sprintf("%s-%s", newSecret.Name, k)] = v
				}
				logrus.Infof("copy secrets %s into %s namespace", secret.Name, settings.IstioExternalLBNamespace)
				if _, err := s.secrets.Create(newSecret); err != nil && !errors.IsAlreadyExists(err) {
					return err
				}
			}
		} else {
			for k, v := range secret.Data {
				exitsingSecret.Data[fmt.Sprintf("%s-%s", secret.Name, k)] = v
			}
			if _, err := s.secrets.Update(exitsingSecret); err != nil {
				return err
			}
		}
	}
	return nil
}
