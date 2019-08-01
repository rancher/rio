package route

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/cli/pkg/tables"
	"github.com/rancher/rio/cli/pkg/types"
	"github.com/urfave/cli"
)

type Ls struct {
}

func (l *Ls) Customize(cmd *cli.Command) {
	cmd.Flags = append(cmd.Flags, table.WriterFlags()...)
}

func (l *Ls) Run(ctx *clicontext.CLIContext) error {
	return ListRouters(ctx)
}

func ListRouters(ctx *clicontext.CLIContext) error {
	routeSets, err := ctx.List(types.RouterType)
	if err != nil {
		return err
	}

	writer := tables.NewRouter(ctx)
	return writer.Write(routeSets)
}
