package cluster

import (
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/urfave/cli"
)

func Cluster(app *cli.App) cli.Command {
	ls := builder.Command(&Ls{},
		"List clusters",
		app.Name+" cluster ls",
		"")
	rm := builder.Command(&Rm{},
		"Remove clusters",
		app.Name+" cluster rm",
		"")
	return cli.Command{
		Name:      "clusters",
		ShortName: "cluster",
		Usage:     "Operations on clusters",
		Action:    clicontext.DefaultAction(ls.Action),
		Flags:     table.WriterFlags(),
		Category:  "SUB COMMANDS",
		Subcommands: []cli.Command{
			ls,
			rm,
		},
	}
}
