package externalservice

import (
	"github.com/rancher/rio/cli/cmd/rm"
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/cli/pkg/tables"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
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
				app.Name+" create [OPTIONS] [EXTERNAL_SERVICE] [(IP)(FQDN)(STACK/SERVICE)]",
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
	Service *riov1.ExternalService
	Stack   *riov1.Stack
}

func externalServiceLs(ctx *clicontext.CLIContext) error {
	externalServices, err := ctx.List(clitypes.ExternalServiceType)
	if err != nil {
		return err
	}

	writer := tables.NewExternalService(ctx)
	return writer.Write(externalServices)
}

func externalServiceRm(ctx *clicontext.CLIContext) error {
	return rm.Remove(ctx, clitypes.ExternalServiceType)
}
