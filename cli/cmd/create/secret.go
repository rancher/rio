package create

import (
	"github.com/rancher/norman/pkg/kv"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

func ParseSecrets(secrets []string) ([]riov1.SecretMapping, error) {
	var result []riov1.SecretMapping
	for _, secret := range secrets {
		mapping, err := parseSecret(secret)
		if err != nil {
			return nil, err
		}
		result = append(result, mapping)
	}

	return result, nil
}

func parseSecret(device string) (riov1.SecretMapping, error) {
	result := riov1.SecretMapping{}

	mapping, optStr := kv.Split(device, ",")
	result.Source, result.Target = kv.Split(mapping, ":")
	opts := kv.SplitMap(optStr, ",")
	result.Mode = opts["mode"]

	return result, nil
}
