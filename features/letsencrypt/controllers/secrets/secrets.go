package secrets

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/rancher/rio/features/letsencrypt/controllers/issuer"
	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3"
	projectv1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/types/apis/core/v1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	syncAllSecrets     = "_secret_all"
	readyAnnotationKey = "certificate-status"
	issuerAnnotation   = "certmanager.k8s.io/issuer-name"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := secretController{
		services:            rContext.Rio.Service,
		secretsLister:       rContext.Core.Secret.Cache(),
		secrets:             rContext.Core.Secret,
		vssController:       rContext.Networking.VirtualService,
		publicDomains:       rContext.Global.PublicDomain,
		publicDomainsLister: rContext.Global.PublicDomain.Cache(),
	}
	rContext.Core.Secret.Interface().AddHandler(ctx, "tls-secrets", c.sync)

	return nil
}

type secretController struct {
	services            riov1.ServiceClient
	secretsLister       v1.SecretClientCache
	secrets             v1.SecretClient
	vssController       v1alpha3.VirtualServiceClient
	publicDomains       projectv1.PublicDomainClient
	publicDomainsLister projectv1.PublicDomainClientCache
}

func (s *secretController) sync(key string, secret *corev1.Secret) (runtime.Object, error) {
	if key == fmt.Sprintf("%s/%s", settings.RioSystemNamespace, syncAllSecrets) {
		return nil, s.syncAllSecrets()
	}
	if secret == nil || secret.DeletionTimestamp != nil || secret.Namespace != settings.RioSystemNamespace {
		return nil, nil
	}
	if secret.Annotations[issuerAnnotation] != "" && len(secret.Data["tls.crt"]) > 0 {
		s.secrets.Enqueue(settings.RioSystemNamespace, syncAllSecrets)
	}
	return nil, nil
}

func (s *secretController) syncAllSecrets() error {
	secrets, err := s.secretsLister.List(settings.RioSystemNamespace, labels.Everything())
	if err != nil {
		return err
	}
	lbSecret, err := s.secretsLister.Get(namespace.CloudNamespace, issuer.TLSSecretName)
	if errors.IsNotFound(err) {
		newSecret := &corev1.Secret{}
		newSecret.Name = issuer.TLSSecretName
		newSecret.Namespace = namespace.CloudNamespace
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
		if secret.Annotations[issuerAnnotation] == "" {
			continue
		}

		if len(secret.Data["tls.crt"]) == 0 {
			secret, err = s.secretsLister.Get(settings.RioSystemNamespace, secret.Name)
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
