package builds

import (
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/urfave/cli"
)

func Builds(app *cli.App) cli.Command {
	ls := builder.Command(&Ls{},
		"List Builds",
		app.Name+" builds ls [OPTIONS] $Name",
		"")
	restart := builder.Command(&Restart{},
		"Restart builds",
		app.Name+" restart $NAMESPACE/$NAME:$revision",
		"")
	return cli.Command{
		Name:      "builds",
		ShortName: "build",
		Usage:     "Operations on builds",
		Action:    clicontext.DefaultAction(ls.Action),
		Flags:     ls.Flags,
		Category:  "SUB COMMANDS",
		Subcommands: []cli.Command{
			ls,
			restart,
		},
	}
}
