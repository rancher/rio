package externalservice

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/rio/cli/cmd/rm"
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	"github.com/urfave/cli"
)

func ExternalService(app *cli.App) cli.Command {
	return cli.Command{
		Name:      "externalservices",
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
	Name   string
	Target string
}

func externalServiceLs(ctx *clicontext.CLIContext) error {
	wc, err := ctx.WorkspaceClient()
	if err != nil {
		return err
	}
	collection, err := wc.ExternalService.List(&types.ListOpts{})
	if err != nil {
		return err
	}
	writer := table.NewWriter([][]string{
		{"NAME", "Name"},
		{"TARGET", "Target"},
	}, ctx)

	defer writer.Close()

	for _, item := range collection.Data {
		writer.Write(&Data{
			Name:   item.Name,
			Target: item.Target,
		})
	}

	return writer.Err()
}

func externalServiceRm(ctx *clicontext.CLIContext) error {
	return rm.Remove(ctx, client.ExternalServiceType)
}
