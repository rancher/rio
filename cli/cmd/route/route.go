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
	create := builder.Command(&Create{},
		"Create a route at the end",
		app.Name+" route create/add MATCH ACTION [TARGET...]",
		"To append a rule at the end, run `rio route add [$NAMESPACE/]$ROUTE_NAME to|redirect|mirror|rewrite [$NAMESPACE/]$SERVICE_NAME")
	create.Aliases = []string{"add"}
	return cli.Command{
		Name:      "routes",
		ShortName: "route",
		Usage:     "Route traffic across the mesh",
		Action:    clicontext.DefaultAction(ls.Action),
		Flags:     table.WriterFlags(),
		Category:  "SUB COMMANDS",
		Subcommands: []cli.Command{
			ls,
			create,
		},
	}
}
