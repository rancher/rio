package publicdomain

import (
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/urfave/cli"
)

func PublicDomain(app *cli.App) cli.Command {
	ls := builder.Command(&Ls{},
		"List public domains",
		app.Name+" domain ls",
		"")
	register := builder.Command(&Register{},
		"Register public domains",
		app.Name+" domain register NAME [NAMESPACE:]SERVICE",
		"Example: run `rio domain register foo.bar [namespace:][service_or_router][@version]`")
	unregister := builder.Command(&Unregister{},
		"Unregister public domains.",
		app.Name+" domain unregister TYPE/NAME",
		"Example: run `rio domain unregister publicdomain/foo.bar`")
	return cli.Command{
		Name:      "domains",
		ShortName: "domain",
		Usage:     "Operations on domains",
		Action:    clicontext.DefaultAction(ls.Action),
		Flags:     ls.Flags,
		Category:  "SUB COMMANDS",
		Subcommands: []cli.Command{
			ls,
			register,
			unregister,
		},
	}
}
