package create

import (
	"github.com/rancher/rio/cli/pkg/kv"
	"github.com/rancher/rio/cli/pkg/kvfile"
)

func parseLabels(files []string, override map[string]string) (map[string]string, error) {
	labels, err := kvfile.ReadKVStrings(files, nil)
	if err != nil {
		return nil, err
	}

	result := map[string]string{}

	for _, label := range labels {
		key, value := kv.Split(label, "=")
		result[key] = value
	}

	for k, v := range override {
		result[k] = v
	}

	return result, nil
}
