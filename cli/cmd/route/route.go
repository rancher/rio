package route

import (
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/urfave/cli"
)

func Route(app *cli.App) cli.Command {
	ls := builder.Command(&Ls{},
		"List routes",
		app.Name+" route ls",
		"")
	return cli.Command{
		Name:      "routes",
		ShortName: "route",
		Usage:     "Route traffic across the mesh",
		Action:    clicontext.DefaultAction(ls.Action),
		Flags:     table.WriterFlags(),
		Subcommands: []cli.Command{
			builder.Command(&Ls{},
				"List routes",
				app.Name+" route ls",
				""),
			builder.Command(&Append{},
				"Append a route at the end",
				app.Name+" route append MATCH ACTION [TARGET...]",
				""),
			builder.Command(&Insert{},
				"Insert a route at the first place",
				app.Name+" route insert MATCH ACTION [TARGET...]",
				""),
		},
	}
}
