package create

import (
	"github.com/rancher/norman/pkg/kv"
	"github.com/rancher/rio/types/client/rio/v1"
)

func ParseSecrets(secrets []string) ([]client.SecretMapping, error) {
	var result []client.SecretMapping
	for _, secret := range secrets {
		mapping, err := parseSecret(secret)
		if err != nil {
			return nil, err
		}
		result = append(result, mapping)
	}

	return result, nil
}

func parseSecret(device string) (client.SecretMapping, error) {
	result := client.SecretMapping{}

	mapping, optStr := kv.Split(device, ",")
	result.Source, result.Target = kv.Split(mapping, ":")
	opts := kv.SplitMap(optStr, ",")
	result.Mode = opts["mode"]

	return result, nil
}
