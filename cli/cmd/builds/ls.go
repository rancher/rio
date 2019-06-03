package builds

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
	cmd.Flags = append(table.WriterFlags(), cmd.Flags...)
}

func (l Ls) Run(ctx *clicontext.CLIContext) error {
	objs, err := ctx.List(types.BuildType)
	if err != nil {
		return err
	}
	writer := tables.NewBuild(ctx)
	return writer.Write(objs)
}
