package config

import (
	"encoding/base64"
	"os"

	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/server"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	"github.com/urfave/cli"
)

type Cat struct {
}

func (c *Cat) Run(app *cli.Context) error {
	ctx, err := server.NewContext(app)
	if err != nil {
		return err
	}
	defer ctx.Close()

	for _, arg := range app.Args() {
		c, err := lookup.Lookup(ctx.Client, arg, client.ConfigType)
		if err != nil {
			return err
		}
		config, err := ctx.Client.Config.ByID(c.ID)
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
