package config

import (
	"fmt"

	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/urfave/cli"
)

func NewCatCommand(sub string, app *cli.App) cli.Command {
	return builder.Command(&Cat{},
		"Print the contents of a config",
		app.Name+sub+" cat [OPTIONS] [NAME...]",
		fmt.Sprintf("To cat all keys, run `rio cat [$NAMESPACE/]$NAME. To cat specific keys, run `rio cat --key foo --key bar [$NAMESPACE/]$NAME"))
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
		Action:    clicontext.DefaultAction(ls.Action),
		Flags:     table.WriterFlags(),
		Subcommands: []cli.Command{
			NewCatCommand(" config", app),
			builder.Command(&Create{},
				"Create a config from a file",
				app.Name+" config create NAME FILE|-",
				""),
			builder.Command(&Ls{},
				"List configs",
				app.Name+" config ls",
				""),
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
