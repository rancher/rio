package externalservice

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/cli/pkg/tables"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	"github.com/urfave/cli"
)

type Ls struct {
}

func (l *Ls) Customize(cmd *cli.Command) {
	cmd.Flags = append(cmd.Flags, table.WriterFlags()...)
}

func (l *Ls) Run(ctx *clicontext.CLIContext) error {
	return ListExternalServices(ctx)
}

func ListExternalServices(ctx *clicontext.CLIContext) error {
	externalServices, err := ctx.List(clitypes.ExternalServiceType)
	if err != nil {
		return err
	}

	writer := tables.NewExternalService(ctx)
	return writer.Write(externalServices)
}
