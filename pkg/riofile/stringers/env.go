package stringers

import (
	"strings"

	"github.com/rancher/rio/cli/pkg/kvfile"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
)

type EnvStringer struct {
	v1.EnvVar
}

func (e *EnvStringer) MaybeString() interface{} {
	buf := &strings.Builder{}
	buf.WriteString(e.Name)
	buf.WriteString("=")
	if e.ConfigMapName != "" {
		buf.WriteString("config://")
		buf.WriteString(e.ConfigMapName)
	} else if e.SecretName != "" {
		buf.WriteString("secret://")
		buf.WriteString(e.SecretName)
	} else {
		buf.WriteString(e.Value)
		return buf.String()
	}

	if e.Key != "" {
		buf.WriteString("/")
		buf.WriteString(e.Key)
	}

	return buf.String()
}

func ParseAllEnv(files []string, envs []string, readEnv bool) (result []v1.EnvVar, err error) {
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

	return ParseEnvs(envs...)
}

func ParseEnvs(envs ...string) (result []v1.EnvVar, err error) {
	for _, env := range envs {
		envVar, err := ParseEnv(env)
		if err != nil {
			return nil, err
		}
		result = append(result, envVar)
	}
	return
}

func ParseEnv(env string) (v1.EnvVar, error) {
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

	return envVar, nil
}

func prefixedKeyValuePath(prefix, value string) (string, string) {
	value = strings.TrimPrefix(value, prefix)
	return kv.Split(value, "/")
}
