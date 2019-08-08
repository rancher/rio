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
	_, err := ListExternalServices(ctx)
	return err
}

func ListExternalServices(ctx *clicontext.CLIContext) (bool, error) {
	externalServices, err := ctx.List(clitypes.ExternalServiceType)
	if err != nil {
		return false, err
	}

	writer := tables.NewExternalService(ctx)
	return len(externalServices) == 0, writer.Write(externalServices)
}
