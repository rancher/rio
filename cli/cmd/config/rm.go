package config

import (
	"github.com/rancher/rio/cli/cmd/rm"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/types"
)

type Rm struct {
}

func (r *Rm) Run(ctx *clicontext.CLIContext) error {
	return rm.Remove(ctx, types.ConfigType)
}
