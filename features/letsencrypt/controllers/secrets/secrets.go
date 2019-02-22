package secrets

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/rancher/rio/features/letsencrypt/controllers/issuer"
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

func (s *secretController) sync(key string, obj *corev1.Secret) (runtime.Object, error) {
	if key == fmt.Sprintf("%s/%s", settings.RioSystemNamespace, syncAllSecrets) {
		return obj, s.syncAllSecrets()
	}
	if obj == nil || obj.DeletionTimestamp != nil || obj.Namespace != settings.RioSystemNamespace {
		return obj, nil
	}
	if obj.Annotations[issuerAnnotation] != "" && len(obj.Data["tls.crt"]) > 0 {
		if obj.Name == issuer.TLSSecretName && obj.Namespace == settings.RioSystemNamespace {
			if err := s.copyBuildCerts(obj); err != nil {
				return nil, err
			}
		}
		s.secrets.Enqueue(settings.RioSystemNamespace, syncAllSecrets)
	}
	return nil, nil
}

func (s *secretController) copyBuildCerts(cert *corev1.Secret) error {
	buildCert := v1.NewSecret(settings.BuildStackName, issuer.TLSSecretName, corev1.Secret{
		Data: cert.Data,
	})
	if _, err := s.secrets.Create(buildCert); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (s *secretController) syncAllSecrets() error {
	secrets, err := s.secretsLister.List(settings.RioSystemNamespace, labels.Everything())
	if err != nil {
		return err
	}

	lbSecret, err := s.secretsLister.Get(settings.IstioStackName, issuer.TLSSecretName)
	if err != nil {
		if errors.IsNotFound(err) {
			newSecret := v1.NewSecret(settings.IstioStackName, issuer.TLSSecretName, corev1.Secret{
				Data: map[string][]byte{},
			})
			created, err := s.secrets.Create(newSecret)
			if err != nil {
				return err
			}
			lbSecret = created
		} else {
			return err
		}
	}

	data := make(map[string][]byte, 0)
	readyDomains := make(map[string]struct{}, 0)
	for _, secret := range secrets {
		if secret.Annotations[issuerAnnotation] == "" {
			continue
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

	if reflect.DeepEqual(data, lbSecret.Data) {
		return nil
	}
	lbSecret.Data = data
	updated, err := s.secrets.Update(lbSecret)
	if err != nil {
		return err
	}

	time.Sleep(time.Minute)

	if updated.Annotations == nil {
		updated.Annotations = make(map[string]string, 0)
	}
	updated.Annotations[readyAnnotationKey] = "ready"
	if _, err = s.secrets.Update(updated); err != nil {
		return err
	}

	for pd := range readyDomains {
		p, err := s.publicDomainsLister.Get(settings.RioSystemNamespace, pd)
		if err != nil && !errors.IsNotFound(err) {
			return err
		} else if errors.IsNotFound(err) {
			continue
		}
		if p.Annotations == nil {
			p.Annotations = make(map[string]string, 0)
		}
		p.Annotations[readyAnnotationKey] = "ready"
		if _, err := s.publicDomains.Update(p); err != nil {
			return err
		}
	}

	logrus.Infof("Certificate %s is updated", lbSecret.Name)
	s.vssController.Enqueue("", "_istio_deploy_")
	return nil
}
