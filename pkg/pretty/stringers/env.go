package stringers

import (
	"strings"

	"github.com/rancher/mapper/mappers"
	"github.com/rancher/rio/cli/pkg/kvfile"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
)

func NewEnv(field string) mappers.ObjectsToSlice {
	return mappers.ObjectsToSlice{
		Field: field,
		NewObject: func() mappers.MaybeStringer {
			return &EnvStringer{}
		},
		ToObject: func(obj interface{}) (interface{}, error) {
			str, ok := obj.(string)
			if !ok {
				return obj, nil
			}
			objs, err := ParseEnv(nil, []string{str}, false)
			if err != nil {
				return nil, err
			}
			return objs[0], nil
		},
	}
}

type EnvStringer struct {
	v1.EnvVar
}

func (e *EnvStringer) MaybeString() interface{} {
	buf := &strings.Builder{}
	buf.WriteString(e.Key)
	buf.WriteString("=")
	if e.SecretName != "" {
		buf.WriteString("secret://")
		buf.WriteString(e.SecretName)
	} else if e.ConfigMapName != "" {
		buf.WriteString("config://")
		buf.WriteString(e.ConfigMapName)
	} else {
		buf.WriteString(e.Value)
	}

	return buf.String()
}

func ParseEnv(files []string, envs []string, readEnv bool) (result []v1.EnvVar, err error) {
	if readEnv {
		envs, err = kvfile.ReadKVEnvStrings(files, envs)
		if err != nil {
			return nil, err
		}
	} else {
		envs, err = kvfile.ReadKVStrings(files, envs)
		if err != nil {
			return nil, err
		}
	}

	for _, env := range envs {
		k, v := kv.Split(env, "=")
		envVar := v1.EnvVar{
			Name:  k,
			Value: v,
		}
		if strings.HasPrefix(v, "secret://") {
			envVar.SecretName, envVar.Key = prefixedKeyValuePath("secret://", v)
		} else if strings.HasPrefix(v, "config://") {
			envVar.ConfigMapName, envVar.Key = prefixedKeyValuePath("config://", v)
		}

		result = append(result, envVar)
	}

	return result, nil
}

func prefixedKeyValuePath(prefix, value string) (string, string) {
	value = strings.TrimPrefix(value, prefix)
	return kv.Split(value, "/")
}
