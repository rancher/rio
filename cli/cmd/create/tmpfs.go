package create

import (
	"fmt"
	"strings"

	units "github.com/docker/go-units"
	"github.com/rancher/norman/pkg/kv"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

func ParseTmpfs(specs []string) ([]riov1.Tmpfs, error) {
	var (
		result []riov1.Tmpfs
		err    error
	)

	for _, spec := range specs {
		var tmpfs riov1.Tmpfs

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
