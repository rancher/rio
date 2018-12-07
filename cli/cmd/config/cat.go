package config

import (
	"encoding/base64"
	"os"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/types/client/rio/v1"
)

type Cat struct {
}

func (c *Cat) Run(ctx *clicontext.CLIContext) error {
	for _, arg := range ctx.CLI.Args() {
		c, err := lookup.Lookup(ctx, arg, client.ConfigType)
		if err != nil {
			return err
		}

		client, err := ctx.ProjectClient()
		if err != nil {
			return err
		}

		config, err := client.Config.ByID(c.ID)
		if err != nil {
			return err
		}

		if len(config.Content) == 0 {
			continue
		}

		var out []byte
		if config.Encoded {
			bytes, err := base64.StdEncoding.DecodeString(config.Content)
			if err != nil {
				return err
			}
			out = bytes
		} else {
			out = []byte(config.Content)
		}

		_, err = os.Stdout.Write(out)
		if err != nil {
			return err
		}
	}

	return nil
}
