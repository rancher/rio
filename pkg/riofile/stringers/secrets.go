package stringers

import (
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

const (
	SecretsDefaultPath = "/run/secrets"
)

type SecretsStringer struct {
	DataMountStringer
}

func (d SecretsStringer) MaybeString() interface{} {
	d.defaultPrefix = SecretsDefaultPath
	return d.DataMountStringer.MaybeString()
}

func ParseSecrets(secrets ...string) (result []v1.DataMount, err error) {
	for _, secret := range secrets {
		s, err := ParseSecret(secret)
		if err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return
}

func ParseSecret(secret string) (v1.DataMount, error) {
	return ParseDataMount(secret)
}
