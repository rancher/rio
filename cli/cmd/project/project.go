package project

import (
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/urfave/cli"
)

func Projects(app *cli.App) cli.Command {
	ls := builder.Command(&Ls{},
		"List projects",
		app.Name+" project ls",
		"")
	add := builder.Command(&Add{},
		"Add projects",
		app.Name+" project add",
		"")
	rm := builder.Command(&Rm{},
		"Rm projects",
		app.Name+" project rm",
		"")
	return cli.Command{
		Name:      "project",
		ShortName: "projects",
		Usage:     "Operations on projects",
		Action:    clicontext.DefaultAction(ls.Action),
		Flags:     ls.Flags,
		Category:  "SUB COMMANDS",
		Subcommands: []cli.Command{
			ls,
			add,
			rm,
		},
	}
}
