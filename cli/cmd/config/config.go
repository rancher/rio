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
		fmt.Sprintf("To cat all keys, run `rio cat [NAMESPACE:]configmap/NAME. To cat specific keys, run `rio cat --key foo --key bar configmap/NAME"))
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
				"Create a config from a file or stdin (with `-` argument)",
				app.Name+" config create [-k KEY] [NAMESPACE:]NAME FILE|-",
				"Example: Set key `hostname` in config map `webapp` to output of command:\n"+
        "         `echo example.com | rio config create -k hostname webapp -`\n"+
        "         Set key `json_config` of config map `app` to content of config.json file:\n"+
        "         `rio config create -k json_config app ./config.json`"),
			builder.Command(&Ls{},
				"List configs",
				app.Name+" config ls",
				""),
			builder.Command(&Rm{},
				"Remove a config",
				app.Name+" config rm [NAMESPACE:][TYPE/NAME...]",
				"Example: run `rio config rm configmap/mysql`"),
			builder.Command(&Update{},
				"Update a config",
				app.Name+" config update [-k KEY] [NAMESPACE:]TYPE/NAME FILE|-",
				"Example: run `rio config update configmap/mysql"),
		},
	}
}
