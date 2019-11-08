package route

import (
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/urfave/cli"
)

func Route(app *cli.App) cli.Command {
	ls := builder.Command(&Ls{},
		"List router",
		app.Name+" router ls",
		"")
	create := builder.Command(&Create{},
		"Create a router at the end",
		app.Name+" router create/add MATCH ACTION [TARGET...]",
		"To append a rule at the end, run `rio [-n $NAMESPACE] router add $ROUTE_NAME to|redirect|mirror|rewrite $SERVICE[@VERSION]")
	create.Aliases = []string{"add"}
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
