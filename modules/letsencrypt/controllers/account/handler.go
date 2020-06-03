package account

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"

	"github.com/go-acme/lego/v3/certcrypto"
	"github.com/go-acme/lego/v3/lego"
	"github.com/go-acme/lego/v3/registration"

	"github.com/rancher/rio/modules/letsencrypt/pkg"

	"github.com/rancher/rio/pkg/config"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/apply"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

var register = false

const (
	handlerName = "letsencrypt-issuer"

	defaultEmail     = "cert@rancher.dev"
	defaultAccount   = "letsencrypt-account"
	defaultServerURL = "https://acme-v02.api.letsencrypt.org/directory"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := &handler{
		key:       fmt.Sprintf("%s/%s", rContext.Namespace, config.ConfigName),
		namespace: rContext.Namespace,
		secrets:   rContext.Core.Core().V1().Secret().Cache(),
		apply: rContext.Apply.
			WithSetID(handlerName).
			WithSetOwnerReference(true, true).
			WithCacheTypes(rContext.Core.Core().V1().Secret()),
	}

	rContext.Core.Core().V1().ConfigMap().OnChange(ctx, handlerName, h.sync)
	return nil
}

type handler struct {
	key       string
	namespace string
	apply     apply.Apply
	secrets   corev1controller.SecretCache
}

func (h *handler) sync(key string, cm *v1.ConfigMap) (*v1.ConfigMap, error) {
	if cm == nil || key != h.key || register {
		return nil, nil
	}

	config, err := config.FromConfigMap(cm)
	if err != nil {
		return cm, err
	}

	if _, err := h.secrets.Get(h.namespace, defaultAccount); err != nil && !errors.IsNotFound(err) {
		return cm, err
	} else if err == nil {
		return cm, nil
	}

	secret, err := constructSecret(h.namespace, config)
	if err != nil {
		return cm, err
	}

	if err := h.apply.WithOwner(cm).ApplyObjects(secret); err != nil {
		return cm, err
	}
	register = true
	return cm, nil
}

func withDefault(val, def string) string {
	if val == "" {
		return def
	}
	return val
}

func constructSecret(namespace string, config config.Config) (*corev1.Secret, error) {
	account := withDefault(config.LetsEncrypt.Account, defaultAccount)
	email := withDefault(config.LetsEncrypt.Email, defaultEmail)
	url := withDefault(config.LetsEncrypt.ServerURL, defaultServerURL)

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	user := &pkg.User{
		Name:  account,
		Email: email,
		URL:   url,
		Key:   privateKey,
	}

	conf := lego.NewConfig(user)

	// This CA URL is configured for a local dev instance of Boulder running in Docker in a VM.
	conf.CADirURL = user.URL
	conf.Certificate.KeyType = certcrypto.RSA2048

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(conf)
	if err != nil {
		return nil, err
	}

	// New users will need to register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, err
	}
	user.Registration = reg

	return pkg.SetSecret(namespace, user)
}
