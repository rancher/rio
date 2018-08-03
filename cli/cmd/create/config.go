package create

import (
	"strconv"

	"github.com/rancher/rio/cli/pkg/kv"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

func ParseConfigs(configs []string) ([]client.ConfigMapping, error) {
	var result []client.ConfigMapping
	for _, config := range configs {
		mapping, err := parseConfig(config)
		if err != nil {
			return nil, err
		}
		result = append(result, mapping)
	}

	return result, nil
}

func parseConfig(device string) (client.ConfigMapping, error) {
	result := client.ConfigMapping{}

	mapping, optStr := kv.Split(device, ",")
	result.Source, result.Target = kv.Split(mapping, ":")
	opts := kv.SplitMap(optStr, ",")

	if i, err := strconv.Atoi(opts["uid"]); err == nil {
		result.UID = int64(i)
	}

	if i, err := strconv.Atoi(opts["gid"]); err == nil {
		result.GID = int64(i)
	}
	result.Mode = opts["mode"]

	return result, nil
}
