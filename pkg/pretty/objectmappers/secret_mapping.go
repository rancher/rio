package objectmappers

import (
	"fmt"

	"github.com/rancher/mapper/mappers"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
)

func NewSecretMapping(field string) mappers.ObjectsToSlice {
	return mappers.ObjectsToSlice{
		Field: field,
		NewObject: func() mappers.MaybeStringer {
			return &SecretMappingStringer{}
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

type SecretMappingStringer struct {
	v1.SecretMapping
}

func (s SecretMappingStringer) MaybeString() interface{} {
	if s.Target == "/"+s.Source {
		s.Target = ""
	}

	msg := s.Source
	if s.Target != "" {
		msg += ":" + s.Target
	}

	if s.Mode != "" {
		msg = fmt.Sprintf("%s,mode=%s", msg, s.Mode)
	}

	return msg
}

func ParseSecrets(secrets ...string) ([]v1.SecretMapping, error) {
	var result []v1.SecretMapping
	for _, secret := range secrets {
		mapping, err := parseSecret(secret)
		if err != nil {
			return nil, err
		}
		result = append(result, mapping)
	}

	return result, nil
}

func parseSecret(device string) (v1.SecretMapping, error) {
	result := v1.SecretMapping{}

	mapping, optStr := kv.Split(device, ",")
	result.Source, result.Target = kv.Split(mapping, ":")
	opts := kv.SplitMap(optStr, ",")
	result.Mode = opts["mode"]

	return result, nil
}
