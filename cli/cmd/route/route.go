package route

import (
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/builder"
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
		Action:    util.DefaultAction(ls.Action),
		Flags:     table.WriterFlags(),
		//Category:  "SUB COMMANDS",
		Subcommands: []cli.Command{
			builder.Command(&Ls{},
				"List routes",
				app.Name+" route ls",
				""),
			builder.Command(&Add{},
				"Add a route",
				app.Name+" route add MATCH ACTION [TARGET...]",
				""),
			//builder.Command(&Rm{},
			//	"Remove a route",
			//	app.Name+" route rm [NAME...]",
			//	""),
		},
	}
}
