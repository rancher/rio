package clicontext

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
)

func (c *CLIContext) Rm(ids ...string) error {
	for _, id := range ids {
		resource, err := c.ByID(id)
		if err != nil && !errors.IsNotFound(err) {
			return err
		}

		if err := c.DeleteResource(resource); err != nil && !errors.IsNotFound(err) {
			return err
		} else if err == nil {
			fmt.Println(id)
		}
	}

	return nil
}
