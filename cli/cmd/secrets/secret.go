package secrets

import (
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/urfave/cli"
)

func Secrets(app *cli.App) cli.Command {
	create := builder.Command(&Create{},
		"Create Secrets",
		app.Name+" secrets create [OPTIONS] $Name",
		"")
	return cli.Command{
		Name:      "secrets",
		ShortName: "secret",
		Usage:     "Operations on secrets",
		Action:    clicontext.DefaultAction(create.Action),
		Flags:     create.Flags,
		Category:  "SUB COMMANDS",
		Subcommands: []cli.Command{
			create,
		},
	}
}
