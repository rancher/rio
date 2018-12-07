package externalservice

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/rio/cli/cmd/rm"
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/types/client/rio/v1"
	"github.com/urfave/cli"
)

func ExternalService(app *cli.App) cli.Command {
	return cli.Command{
		Name:      "externalservices",
		Aliases:   []string{"external"},
		ShortName: "externalservice",
		Usage:     "Operation on externalservices",
		Action:    clicontext.DefaultAction(externalServiceLs),
		Flags:     table.WriterFlags(),
		Category:  "SUB COMMANDS",
		Subcommands: []cli.Command{
			{
				Name:      "ls",
				Usage:     "List external services",
				ArgsUsage: "None",
				Action:    clicontext.Wrap(externalServiceLs),
				Flags:     table.WriterFlags(),
			},
			builder.Command(&Create{},
				"Create external services",
				app.Name+" create [OPTIONS] [EXTERNAL_SERVICE...]",
				""),
			{
				Name:      "delete",
				ShortName: "rm",
				Usage:     "Delete a stack",
				ArgsUsage: "None",
				Action:    clicontext.Wrap(externalServiceRm),
			},
		},
	}
}

type Data struct {
	Name    string
	Target  string
	Created string
	Service *client.ExternalService
	Stack   *client.Stack
}

func externalServiceLs(ctx *clicontext.CLIContext) error {
	wc, err := ctx.ProjectClient()
	if err != nil {
		return err
	}
	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}
	collection, err := wc.ExternalService.List(&types.ListOpts{})
	if err != nil {
		return err
	}
	writer := table.NewWriter([][]string{
		{"NAME", "{{stackScopedName .Stack.Name .Service.Name}}"},
		{"CREATED", "{{.Created | ago}}"},
		{"TARGET", "{{.Service.Target}}"},
	}, ctx)
	writer.AddFormatFunc("stackScopedName", table.FormatStackScopedName(cluster))
	defer writer.Close()

	stackByID, err := util.StacksByID(wc)
	if err != nil {
		return err
	}

	for _, item := range collection.Data {
		writer.Write(&Data{
			Name:    item.Name,
			Target:  item.Target,
			Created: item.Created,
			Stack:   stackByID[item.StackID],
			Service: &item,
		})
	}

	return writer.Err()
}

func externalServiceRm(ctx *clicontext.CLIContext) error {
	return rm.Remove(ctx, client.ExternalServiceType)
}
