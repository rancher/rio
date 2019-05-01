package stringers

import (
	"github.com/rancher/mapper/mappers"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

const (
	secretsDefaultPath = "/run/secrets"
)

func NewSecrets(field string) mappers.ObjectsToSlice {
	return mappers.ObjectsToSlice{
		Field: field,
		NewObject: func() mappers.MaybeStringer {
			return &SecretsStringer{}
		},
		ToObject: func(obj interface{}) (interface{}, error) {
			str, ok := obj.(string)
			if !ok {
				return obj, nil
			}
			objs, err := ParseSecrets(str)
			if err != nil {
				return nil, err
			}
			return objs[0], nil
		},
	}
}

type SecretsStringer struct {
	DataMountStringer
}

func (d SecretsStringer) MaybeString() interface{} {
	d.defaultPrefix = secretsDefaultPath
	return d.DataMountStringer.MaybeString()
}

func ParseSecrets(secrets ...string) ([]v1.DataMount, error) {
	return ParseDataMounts(secretsDefaultPath, secrets...)
}
