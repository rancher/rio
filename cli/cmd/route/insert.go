package route

import "github.com/rancher/rio/cli/pkg/clicontext"

type Insert struct {
	Add
}

func (i *Insert) Run(ctx *clicontext.CLIContext) error {
	return insertRoute(ctx, true, i)
}
