package builds

import (
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/urfave/cli"
)

func Builds(app *cli.App) cli.Command {
	ls := builder.Command(&History{},
		"List Builds",
		app.Name+" builds history [OPTIONS]",
		"")
	return cli.Command{
		Name:     "build",
		Usage:    "Operations on builds",
		Flags:    ls.Flags,
		Category: "SUB COMMANDS",
		Subcommands: []cli.Command{
			ls,
		},
	}
}
