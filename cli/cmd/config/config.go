package config

import (
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/urfave/cli"
)

func NewCatCommand(sub string, app *cli.App) cli.Command {
	return builder.Command(&Cat{},
		"Print the contents of a config",
		app.Name+sub+" cat [NAME...]",
		"")
}

func Config(app *cli.App) cli.Command {
	ls := builder.Command(&Ls{},
		"List configs",
		app.Name+" config ls",
		"")
	return cli.Command{
		Name:      "configs",
		ShortName: "config",
		Usage:     "Operations on configs",
		Category:  "SUB COMMANDS",
		Action:    util.DefaultAction(ls.Action),
		Flags:     table.WriterFlags(),
		Subcommands: []cli.Command{
			builder.Command(&Ls{},
				"List configs",
				app.Name+" config ls",
				""),
			builder.Command(&Create{},
				"Create a config",
				app.Name+" config create NAME FILE|-",
				""),
			NewCatCommand(" config", app),
			builder.Command(&Rm{},
				"Remove a config",
				app.Name+" config rm [NAME...]",
				""),
			builder.Command(&Update{},
				"Update a config",
				app.Name+" config update NAME FILE|-",
				""),
		},
	}
}
