package config

import (
	"encoding/base64"
	"os"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/types"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

type Cat struct {
}

func (c *Cat) Run(ctx *clicontext.CLIContext) error {
	for _, arg := range ctx.CLI.Args() {
		r, err := lookup.Lookup(ctx, arg, types.ConfigType)
		if err != nil {
			return err
		}

		config := r.Object.(*v1.Config)

		if len(config.Spec.Content) == 0 {
			continue
		}

		var out []byte
		if config.Spec.Encoded {
			bytes, err := base64.StdEncoding.DecodeString(config.Spec.Content)
			if err != nil {
				return err
			}
			out = bytes
		} else {
			out = []byte(config.Spec.Content)
		}

		_, err = os.Stdout.Write(out)
		if err != nil {
			return err
		}
	}

	return nil
}
