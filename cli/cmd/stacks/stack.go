package stacks

import (
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/urfave/cli"
)

func Stacks(app *cli.App) cli.Command {
	ls := builder.Command(&ls{},
		"List stacks",
		app.Name+" stack ls [OPTIONS] $Name",
		"")
	info := builder.Command(&info{},
		"describe general information for stack",
		app.Name+" stack info $NAME",
		"")
	update := builder.Command(&update{},
		"Update stack answers and images",
		app.Name+" stack cat $Name",
		"To update answer file for a stack, run `rio stack update --answers $FILE_PATH $NAME`. To update images for services in stack, run `rio stack update --images service-foo=nginx $NAME`")
	return cli.Command{
		Name:      "stacks",
		ShortName: "stack",
		Usage:     "Operations on stack",
		Action:    clicontext.DefaultAction(ls.Action),
		Flags:     ls.Flags,
		Category:  "SUB COMMANDS",
		Subcommands: []cli.Command{
			ls,
			update,
			info,
		},
	}
}
