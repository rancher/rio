package create

import (
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/pretty/stringers"
)

func populateMemory(c *Create, service *riov1.Service) error {
	if c.M_Memory != "" {
		memory, err := stringers.ParseQuantity(c.M_Memory)
		if err != nil {
			return err
		}
		service.Spec.Memory = &memory
	}

	return nil
}
