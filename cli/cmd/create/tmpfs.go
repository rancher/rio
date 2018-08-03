package create

import (
	"fmt"
	"strings"

	"github.com/docker/go-units"
	"github.com/rancher/rio/cli/pkg/kv"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

func ParseTmpfs(specs []string) ([]client.Tmpfs, error) {
	var (
		result []client.Tmpfs
		err    error
	)

	for _, spec := range specs {
		var tmpfs client.Tmpfs

		name, opts := kv.Split(spec, ":")
		for _, opt := range strings.Split(opts, ",") {
			key, value := kv.Split(opt, "=")
			switch key {
			case "ro":
				tmpfs.ReadOnly = true
			case "size":
				tmpfs.SizeBytes, err = units.RAMInBytes(value)
				if err != nil {
					return nil, fmt.Errorf("failed to parse %s: %v", opt, err)
				}
			}
		}

		tmpfs.Path = name
		result = append(result, tmpfs)
	}

	return result, nil
}
