package volume

import (
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/urfave/cli"
)

func Volume(app *cli.App) cli.Command {
	ls := builder.Command(&Ls{},
		"List volumes",
		app.Name+" volume ls",
		"")
	return cli.Command{
		Name:      "volumes",
		ShortName: "volume",
		Usage:     "Operations on volumes",
		Action:    clicontext.DefaultAction(ls.Action),
		Flags:     table.WriterFlags(),
		Category:  "SUB COMMANDS",
		Subcommands: []cli.Command{
			ls,
			builder.Command(&Create{},
				"Create a volume",
				app.Name+" volume create NAME SIZE_IN_GB",
				""),
			builder.Command(&Rm{},
				"Remove a volume",
				app.Name+" volume rm [NAME...]",
				""),
		},
	}
}
