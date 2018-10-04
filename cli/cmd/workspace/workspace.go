package workspace

import (
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/urfave/cli"
)

func Workspace(app *cli.App) cli.Command {
	ls := builder.Command(&Ls{},
		"List workspaces",
		app.Name+" workspace ls",
		"")
	return cli.Command{
		Name:      "workspaces",
		ShortName: "workspace",
		Usage:     "Operations on workspaces",
		Action:    clicontext.DefaultAction(ls.Action),
		Flags:     ls.Flags,
		Category:  "SUB COMMANDS",
		Subcommands: []cli.Command{
			ls,
		},
	}
}
