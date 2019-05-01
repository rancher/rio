package create

import (
	"fmt"

	units "github.com/docker/go-units"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

func populateMemory(c *Create, service *riov1.Service) error {
	var err error

	if c.M_Memory != "" {
		service.Spec.MemoryReservationBytes, err = units.RAMInBytes(c.M_Memory)
		if err != nil {
			return fmt.Errorf("failed to parse --memory-reservation: %v", err)
		}
	}

	if c.MemoryLimit != "" {
		service.Spec.MemoryLimitBytes, err = units.RAMInBytes(c.MemoryLimit)
		if err != nil {
			return fmt.Errorf("failed to parse --memory-limit: %v", err)
		}
	}

	return nil
}
