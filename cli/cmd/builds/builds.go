package builds

import (
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/urfave/cli"
)

func Builds(app *cli.App) cli.Command {
	ls := builder.Command(&History{},
		"List Builds",
		app.Name+" builds history [OPTIONS]",
		"")
	restart := builder.Command(&Restart{},
		"Restart builds",
		app.Name+" restart $NAMESPACE/$NAME:$revision",
		"To restart a build, run `rio restart default/build-foo:bar")
	return cli.Command{
		Name:     "build",
		Usage:    "Operations on builds",
		Flags:    ls.Flags,
		Category: "SUB COMMANDS",
		Subcommands: []cli.Command{
			ls,
			restart,
		},
	}
}
