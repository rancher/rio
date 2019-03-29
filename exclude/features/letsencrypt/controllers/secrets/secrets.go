package secrets

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rancher/rio/features/letsencrypt/controllers/issuer"
	"github.com/rancher/rio/features/routing/controllers/istio"
	"github.com/rancher/rio/pkg/constructors"
	v12 "github.com/rancher/rio/pkg/generated/controllers/core/v1"
	"github.com/rancher/rio/pkg/generated/controllers/networking.istio.io/v1alpha3"
	v13 "github.com/rancher/rio/pkg/generated/controllers/project.rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/trigger"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	readyAnnotationKey = "certificate-status"
	issuerAnnotation   = "certmanager.k8s.io/issuer-name"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := secretController{
		systemNamespace:     rContext.SystemNamespace,
		services:            rContext.Rio.Rio().V1().Service(),
		secretsLister:       rContext.Core.Core().V1().Secret().Cache(),
		secrets:             rContext.Core.Core().V1().Secret(),
		vssController:       rContext.Networking.Networking().V1alpha3().VirtualService(),
		publicDomains:       rContext.Global.Project().V1().PublicDomain(),
		publicDomainsLister: rContext.Global.Project().V1().PublicDomain().Cache(),
		trigger:             trigger.New(rContext.Core.Core().V1().Secret()),
	}

	c.trigger.OnTrigger(ctx, "tls-secrets", c.syncAllSecrets)
	rContext.Core.Core().V1().Secret().OnChange(ctx, "tls-secrets", c.sync)

	return nil
}

type secretController struct {
	systemNamespace     string
	services            v1.ServiceClient
	secretsLister       v12.SecretCache
	secrets             v12.SecretClient
	vssController       v1alpha3.VirtualServiceController
	publicDomains       v13.PublicDomainClient
	publicDomainsLister v13.PublicDomainCache
	trigger             trigger.Trigger
}

func (s *secretController) sync(key string, obj *corev1.Secret) (*corev1.Secret, error) {
	if obj == nil || obj.DeletionTimestamp != nil || obj.Namespace != s.systemNamespace {
		return obj, nil
	}

	if obj.Annotations[issuerAnnotation] != "" && len(obj.Data["tls.crt"]) > 0 {
		if obj.Name == issuer.TLSSecretName && obj.Namespace == s.systemNamespace {
			if err := s.copyBuildCerts(obj); err != nil {
				return nil, err
			}
		}
		s.trigger.Trigger()
	}

	return nil, nil
}

func (s *secretController) copyBuildCerts(cert *corev1.Secret) error {
	buildCert := constructors.NewSecret(settings.BuildStackName, issuer.TLSSecretName, corev1.Secret{
		Data: cert.Data,
	})
	if _, err := s.secrets.Create(buildCert); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (s *secretController) syncAllSecrets() error {
	secrets, err := s.secretsLister.List(s.systemNamespace, labels.Everything())
	if err != nil {
		return err
	}

	lbSecret, err := s.secretsLister.Get(settings.IstioStackName, issuer.TLSSecretName)
	if errors.IsNotFound(err) {
		newSecret := constructors.NewSecret(settings.IstioStackName, issuer.TLSSecretName, corev1.Secret{
			Data: map[string][]byte{},
		})
		lbSecret, err = s.secrets.Create(newSecret)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	data := map[string][]byte{}
	readyDomains := map[string]struct{}{}

	for _, secret := range secrets {
		if secret.Annotations[issuerAnnotation] == "" {
			continue
		}

		if len(secret.Data["tls.crt"]) <= 0 {
			continue
		}

		if strings.HasSuffix(secret.Name, "-tls-certs") {
			readyDomains[strings.TrimSuffix(secret.Name, "-tls-certs")] = struct{}{}
		}
		for k, v := range secret.Data {
			data[fmt.Sprintf("%s-%s", secret.Name, k)] = v
		}
	}

	if equality.Semantic.DeepEqual(data, lbSecret.Data) {
		return nil
	}

	lbSecret = lbSecret.DeepCopy()
	lbSecret.Data = data
	updated, err := s.secrets.Update(lbSecret)
	if err != nil {
		return err
	}

	// TODO: I am a terrible person
	time.Sleep(time.Minute)

	if updated.Annotations == nil {
		updated.Annotations = map[string]string{}
	}

	updated.Annotations[readyAnnotationKey] = "ready"
	if _, err = s.secrets.Update(updated); err != nil {
		return err
	}

	for pd := range readyDomains {
		p, err := s.publicDomainsLister.Get(s.systemNamespace, pd)
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
	istio.ReevalIstio()
	return nil
}
