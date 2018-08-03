package config

import (
	"github.com/rancher/rio/cli/cmd/rm"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	"github.com/urfave/cli"
)

type Rm struct {
}

func (r *Rm) Run(app *cli.Context) error {
	return rm.Remove(app, client.ConfigType)
}
