package buildhistory

import (
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/cli/pkg/tables"
	"github.com/rancher/rio/cli/pkg/types"
	"github.com/urfave/cli"
)

func History(app *cli.App) cli.Command {
	hist := builder.Command(&BuildHistory{},
		"List Build History",
		app.Name+" build-history [OPTIONS]",
		"")
	return cli.Command{
		Name:   "build-history",
		Usage:  "Show build-history",
		Action: clicontext.DefaultAction(hist.Action),
		Flags:  table.WriterFlags(),
	}
}

type BuildHistory struct {
}

func (b *BuildHistory) Customize(cmd *cli.Command) {
	cmd.Flags = append(table.WriterFlags(), cmd.Flags...)
}

func (b BuildHistory) Run(ctx *clicontext.CLIContext) error {
	objs, err := ctx.List(types.BuildType)
	if err != nil {
		return err
	}
	writer := tables.NewBuild(ctx)
	return writer.Write(objs)
}
