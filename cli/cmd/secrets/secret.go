package secrets

import (
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/urfave/cli"
)

func Secrets(app *cli.App) cli.Command {
	create := builder.Command(&Create{},
		"Create Secrets",
		app.Name+" secrets create [OPTIONS] NAME",
		"")
	create.Aliases = []string{"add"}
	ls := builder.Command(&Ls{},
		"List Secrets",
		app.Name+" secrets ls",
		"")
	return cli.Command{
		Name:      "secrets",
		ShortName: "secret",
		Usage:     "Operations on secrets",
		Action:    clicontext.DefaultAction(ls.Action),
		Flags:     create.Flags,
		Category:  "SUB COMMANDS",
		Subcommands: []cli.Command{
			ls,
			create,
		},
	}
}
