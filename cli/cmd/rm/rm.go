package rm

import (
	"errors"

	"github.com/rancher/rio/cli/pkg/clicontext"
)

type Rm struct {
	T_Type string `desc:"delete specific type. Available types: [config,service,router,externalservice,publicdomain,app,secret,build,stack]"`
}

func (r *Rm) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return errors.New("at least one argument is needed")
	}

	return ctx.Rm(ctx.CLI.Args()...)
}
