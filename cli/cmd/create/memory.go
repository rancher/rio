package create

import (
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/riofile/stringers"
)

func (c *Create) setMemory(spec *riov1.ServiceSpec) error {
	if c.M_Memory != "" {
		memory, err := stringers.ParseQuantity(c.M_Memory)
		if err != nil {
			return err
		}
		m := memory.AsDec().UnscaledBig().Int64()
		spec.MemoryBytes = &m
	}

	return nil
}
