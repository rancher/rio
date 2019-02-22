package create

import (
	"strconv"

	"github.com/rancher/norman/pkg/kv"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

func ParseConfigs(configs []string) ([]riov1.ConfigMapping, error) {
	var result []riov1.ConfigMapping
	for _, config := range configs {
		mapping, err := parseConfig(config)
		if err != nil {
			return nil, err
		}
		result = append(result, mapping)
	}

	return result, nil
}

func parseConfig(device string) (riov1.ConfigMapping, error) {
	result := riov1.ConfigMapping{}

	mapping, optStr := kv.Split(device, ",")
	result.Source, result.Target = kv.Split(mapping, ":")
	opts := kv.SplitMap(optStr, ",")

	if i, err := strconv.Atoi(opts["uid"]); err == nil {
		result.UID = i
	}

	if i, err := strconv.Atoi(opts["gid"]); err == nil {
		result.GID = i
	}
	result.Mode = opts["mode"]

	return result, nil
}
