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
	add := builder.Command(&Add{},
		"List public domains",
		app.Name+" domain add $Name",
		"")
	rm := builder.Command(&Rm{},
		"List public domains",
		app.Name+" domain rm NAME",
		"")
	return cli.Command{
		Name:      "domains",
		ShortName: "domain",
		Usage:     "Operations on domains",
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
