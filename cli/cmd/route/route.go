package route

import (
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/urfave/cli"
)

func Route(app *cli.App) cli.Command {
	create := builder.Command(&Create{},
		"Create a router at the end",
		app.Name+" router create/add MATCH ACTION [TARGET...]",
		"To append a rule at the end, run `rio router add $ROUTE_NAME to|redirect|mirror|rewrite $APP[@VERSION]`. If version not specified app is assumed.")
	create.Aliases = []string{"add"}
	ls := builder.Command(&Ls{},
		"List router",
		app.Name+" router ls",
		"")
	return cli.Command{
		Name:     "routers",
		Aliases:  []string{"router", "route"},
		Usage:    "Route traffic across the mesh",
		Action:   clicontext.DefaultAction(ls.Action),
		Flags:    table.WriterFlags(),
		Category: "SUB COMMANDS",
		Subcommands: []cli.Command{
			ls,
			create,
		},
	}
}
