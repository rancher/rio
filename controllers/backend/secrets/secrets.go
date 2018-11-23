package secrets

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/rancher/rio/pkg/certs"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	spacev1beta1 "github.com/rancher/rio/types/apis/space.cattle.io/v1beta1"
	"github.com/rancher/types/apis/core/v1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	syncAllSecrets     = "_secret_all"
	readyAnnotationKey = "certificate-status"
)

func Register(ctx context.Context, rContext *types.Context) {
	c := secretController{
		services:            rContext.Rio.Services(""),
		secretsLister:       rContext.Core.Secrets("").Controller().Lister(),
		secrets:             rContext.Core.Secrets(""),
		secretRefresher:     rContext.Core.Secrets(settings.RioSystemNamespace),
		secretsController:   rContext.Core.Secrets(settings.RioSystemNamespace).Controller(),
		vssController:       rContext.Networking.VirtualServices("").Controller(),
		publicDomains:       rContext.Global.PublicDomains(settings.RioSystemNamespace),
		publicDomainsLister: rContext.Global.PublicDomains(settings.RioSystemNamespace).Controller().Lister(),
	}
	rContext.Core.Secrets(settings.RioSystemNamespace).AddHandler(ctx, "tls-secrets", c.sync)
}

type secretController struct {
	services            v1beta1.ServiceInterface
	secretsLister       v1.SecretLister
	secretRefresher     v1.SecretInterface
	secrets             v1.SecretInterface
	secretsController   v1.SecretController
	vssController       v1alpha3.VirtualServiceController
	publicDomains       spacev1beta1.PublicDomainInterface
	publicDomainsLister spacev1beta1.PublicDomainLister
}

func (s *secretController) sync(key string, secret *corev1.Secret) (runtime.Object, error) {
	if key == fmt.Sprintf("%s/%s", settings.RioSystemNamespace, syncAllSecrets) {
		return nil, s.syncAllSecrets()
	}
	if secret == nil || secret.DeletionTimestamp != nil {
		return nil, nil
	}
	if secret.Annotations["certmanager.k8s.io/issuer-name"] == settings.CerManagerIssuerName && len(secret.Data["tls.crt"]) > 0 {
		s.secretsController.Enqueue(settings.RioSystemNamespace, syncAllSecrets)
	}
	return nil, nil
}

func (s *secretController) syncAllSecrets() error {
	secrets, err := s.secretsLister.List(settings.RioSystemNamespace, labels.Everything())
	if err != nil {
		return err
	}
	lbSecret, err := s.secretsLister.Get(settings.IstioExternalLBNamespace, certs.TlsSecretName)
	if errors.IsNotFound(err) {
		newSecret := &corev1.Secret{}
		newSecret.Name = certs.TlsSecretName
		newSecret.Namespace = settings.IstioExternalLBNamespace
		newSecret.Data = make(map[string][]byte)
		created, err := s.secrets.Create(newSecret)
		if err != nil {
			return err
		}
		lbSecret = created
	} else if err != nil {
		return err
	}
	updateCopy := lbSecret.DeepCopy()
	data := make(map[string][]byte, 0)
	readyDomains := make(map[string]struct{}, 0)
	for _, secret := range secrets {
		if secret.Annotations["certmanager.k8s.io/issuer-name"] == settings.CerManagerIssuerName {
			if len(secret.Data["tls.crt"]) == 0 {
				secret, err = s.secretRefresher.Get(secret.Name, metav1.GetOptions{})
				if err != nil {
					return err
				}
			}
			if len(secret.Data["tls.crt"]) > 0 {
				if strings.HasSuffix(secret.Name, "-tls-certs") {
					readyDomains[strings.TrimSuffix(secret.Name, "-tls-certs")] = struct{}{}
				}
				for k, v := range secret.Data {
					data[fmt.Sprintf("%s-%s", secret.Name, k)] = v
				}
			}
		}
	}

	if reflect.DeepEqual(data, updateCopy.Data) {
		return nil
	}
	updateCopy.Data = data

	updated, err := s.secrets.Update(updateCopy)
	if err != nil {
		return err
	}
	time.Sleep(time.Minute)
	for pd := range readyDomains {
		p, err := s.publicDomainsLister.Get(settings.RioSystemNamespace, pd)
		if err != nil && !errors.IsNotFound(err) {
			return err
		} else if errors.IsNotFound(err) {
			continue
		}
		deepCp := p.DeepCopy()
		if deepCp.Annotations == nil {
			deepCp.Annotations = make(map[string]string, 0)
		}
		deepCp.Annotations[readyAnnotationKey] = "ready"
		if _, err := s.publicDomains.Update(deepCp); err != nil {
			return err
		}
	}
	if updated.Annotations == nil {
		updated.Annotations = make(map[string]string, 0)
	}
	updated.Annotations[readyAnnotationKey] = "ready"
	_, err = s.secrets.Update(updated)
	if err != nil {
		return err
	}
	logrus.Infof("Certificate %s is updated", updateCopy.Name)
	s.vssController.Enqueue("", "_istio_deploy_")
	return nil
}
